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
	"sync"

	bq "cloud.google.com/go/bigquery"

	"github.com/spf13/cobra"
	mpb "github.com/vbauerster/mpb/v8"
	decor "github.com/vbauerster/mpb/v8/decor"
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
	ctx := context.Background()
	var metas metadata.Metas

	projects, err := listProjects(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch GCP projects: %s", err)
	}

	fatalErrors := make(chan error)
	wgDone := make(chan bool)

	var mutex sync.Mutex
	var wg sync.WaitGroup
	pb := mpb.NewWithContext(ctx, mpb.WithWidth(32), mpb.WithWaitGroup(&wg))

	for i, p := range *projects {
		i := i
		p := p

		ds, err := listDataSets(ctx, p)
		if err != nil {
			return fmt.Errorf("failed to fetch BigQuery datasets: project %s, error %s", p, err)
		}

		bar := NewBar(pb, int64(len(*ds)), fmt.Sprintf("[%v/%v][%v] caching datasets...", i+1, len(*projects), p))

		go func() {
			client, err := bq.NewClient(ctx, p)
			if err != nil {
				fatalErrors <- err
			}
			defer client.Close()

			for _, d := range *ds {
				projectMetas, err := listMetaData(ctx, client, p, d)
				if err != nil {
					err = fmt.Errorf("failed to fetch metadata: project %s, error %s", p, err)
					fatalErrors <- err
				}
				mutex.Lock()
				metas.Metas = append(metas.Metas, projectMetas.Metas...)
				mutex.Unlock()
				bar.Increment()
			}
		}()
	}

	go func() {
		pb.Wait()
		close(wgDone)
	}()

	select {
	case <-wgDone:
		break // carry on
	case err := <-fatalErrors:
		close(fatalErrors)
		return err
	}

	err = metas.Save(config.CacheFile)
	if err != nil {
		return fmt.Errorf("failed to save cache: %s", err)
	}

	fmt.Printf("dataset meta data are cached to %s\n", config.CacheFile)
	return nil
}

func listProjects(ctx context.Context) (*[]string, error) {
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

func listDataSets(ctx context.Context, project string) (*[]string, error) {
	client, err := bq.NewClient(ctx, project)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	it := client.Datasets(ctx)
	var datasets []string
	for {
		ds, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to fetch dataset: %w", err)
		}
		datasets = append(datasets, ds.DatasetID)
	}
	return &datasets, nil
}

func listMetaData(ctx context.Context, client *bq.Client, project, dataset string) (metadata.Metas, error) {
	md, err := client.Dataset(dataset).Metadata(ctx)
	if err != nil {
		return metadata.Metas{}, fmt.Errorf("failed to fetch dataset metadata: project %s, dataset %s: %w", project, dataset, err)
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

func NewBar(pb *mpb.Progress, max int64, name string) *mpb.Bar {
	return pb.AddBar(max,
		mpb.PrependDecorators(
			decor.Name(name),
			decor.Percentage(decor.WCSyncSpace),
		),
		mpb.AppendDecorators(
			decor.OnComplete(
				decor.Elapsed(decor.ET_STYLE_GO, decor.WCSyncSpace), "done",
			),
		),
	)
}

func init() {
	rootCmd.AddCommand(cacheCmd)
}
