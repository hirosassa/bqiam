/*
Copyright © 2020 Hirohito Sasakawa

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"context"
	"errors"
	"fmt"

	bq "cloud.google.com/go/bigquery"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"google.golang.org/api/bigquery/v2"
	"google.golang.org/api/iterator"

	"github.com/hirosassa/bqiam/metadata"
)

// cacheCmd represents the cache command
var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "cache fetches bigquery datasets' metadata and store it in local file",
	Long:  `cache fetches bigquery datasets' metadata and store it in local file`,
	RunE:  runCmdCache,
}

func runCmdCache(cmd *cobra.Command, args []string) error {
	var metas metadata.Metas

	projects, err := listProjects()
	if err != nil {
		return fmt.Errorf("failed to fetch GCP projects: %s", err)
	}

	n := len(*projects)
	for i, p := range *projects {
		ds, err := listDataSets(p)
		if err != nil {
			return fmt.Errorf("failed to fetch BigQuery datasets: project %s, error %s", p, err)
		}

		bar := newProgressbar(fmt.Sprintf("[%v/%v][%v] caching Datasets...", i+1, n, p), len(*ds))

		for _, d := range *ds {
			projectMetas, err := listMetaData(p, d)
			if err != nil {
				return fmt.Errorf("failed to fetch metadata: project %s, error %s", p, err)
			}
			metas.Metas = append(metas.Metas, projectMetas.Metas...)
			bar.Add(1)
		}
		fmt.Println("  done!")
	}

	err = metas.Save(config.CacheFile)
	if err != nil {
		return fmt.Errorf("failed to save cache: %s", err)
	}

	fmt.Printf("dataset meta data are cached to %s\n", config.CacheFile)
	return nil
}

func listProjects() (*[]string, error) {
	ctx := context.Background()
	bigqueryService, err := bigquery.NewService(ctx)
	if err != nil {
		return nil, errors.New("failed to create bigqueryService")
	}

	var pageToken string
	var projects []string
	for {
		call := bigqueryService.Projects.List()
		if len(pageToken) > 0 {
			call = call.PageToken(pageToken)
		}

		list, err := call.Do()
		if err != nil {
			return nil, errors.New("failed to call bigquery API")
		}

		for _, project := range list.Projects { // extract bigquery project
			if isBigQueryProject(project.Id) {
				projects = append(projects, project.Id)
			}
		}

		pageToken = list.NextPageToken
		if len(pageToken) == 0 {
			break
		}
	}
	return &projects, nil
}

func listDataSets(project string) (*[]string, error) {
	ctx := context.Background()
	client, err := bq.NewClient(ctx, project)
	if err != nil {
		return nil, errors.New("failed to create client")
	}

	it := client.Datasets(ctx)
	var datasets []string
	for {
		ds, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, errors.New("failed to fetch dataset")
		}
		datasets = append(datasets, ds.DatasetID)
	}
	return &datasets, nil
}

func listMetaData(project, dataset string) (metadata.Metas, error) {
	ctx := context.Background()
	client, err := bq.NewClient(ctx, project)
	if err != nil {
		return metadata.Metas{}, errors.New("failed to create client")
	}
	md, err := client.Dataset(dataset).Metadata(ctx)
	if err != nil {
		return metadata.Metas{}, fmt.Errorf("failed to fetch dataset metadata: project %s,  dataset %s", project, dataset)
	}

	var metas metadata.Metas
	for _, a := range md.Access {
		d := metadata.Meta{
			Project: project,
			Dataset: dataset,
			Role:    a.Role,
			Entity:  a.Entity,
		}
		metas.Metas = append(metas.Metas, d)
	}
	return metas, nil
}

func isBigQueryProject(project string) bool {
	bp := config.BigqueryProjects
	for _, p := range bp {
		if project == p {
			return true
		}
	}
	return false
}

func init() {
	rootCmd.AddCommand(cacheCmd)
}

func newProgressbar(description string, max int) *progressbar.ProgressBar {
	return progressbar.NewOptions(
		max,
		progressbar.OptionSetWidth(20),
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)
}
