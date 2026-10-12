package cmd

import (
	"context"

	pb "github.com/infinimesh/proto/node"
	accpb "github.com/infinimesh/proto/node/accounts"
	devpb "github.com/infinimesh/proto/node/devices"
	nspb "github.com/infinimesh/proto/node/namespaces"
)

// The API returns at most one page per call, so these follow next_page_token to the end.

func listAllAccounts(ctx context.Context, c pb.AccountsServiceClient) (*accpb.Accounts, error) {
	all := &accpb.Accounts{}
	token := ""
	for {
		r, err := c.List(ctx, &pb.ListRequest{PageToken: token})
		if err != nil {
			return nil, err
		}
		all.Accounts = append(all.Accounts, r.GetAccounts()...)
		if token = r.GetNextPageToken(); token == "" {
			return all, nil
		}
	}
}

func listAllNamespaces(ctx context.Context, c pb.NamespacesServiceClient) (*nspb.Namespaces, error) {
	all := &nspb.Namespaces{}
	token := ""
	for {
		r, err := c.List(ctx, &pb.ListRequest{PageToken: token})
		if err != nil {
			return nil, err
		}
		all.Namespaces = append(all.Namespaces, r.GetNamespaces()...)
		if token = r.GetNextPageToken(); token == "" {
			return all, nil
		}
	}
}

func listAllDevices(ctx context.Context, c pb.DevicesServiceClient, q *pb.QueryRequest) (*devpb.Devices, error) {
	all := &devpb.Devices{}
	for {
		r, err := c.List(ctx, q)
		if err != nil {
			return nil, err
		}
		all.Devices = append(all.Devices, r.GetDevices()...)
		if q.PageToken = r.GetNextPageToken(); q.PageToken == "" {
			return all, nil
		}
	}
}
