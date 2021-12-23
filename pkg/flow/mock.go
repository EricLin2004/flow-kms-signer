package flow

import (
	"context"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/crypto"
	"google.golang.org/grpc"
)

type MockClient struct {
}

type mockImpl struct {
	client *MockClient
}

func (m *mockImpl) GetFlowClient() *client.Client {
	return nil
}

func (m *mockImpl) SendSignedTransaction(ctx context.Context, signedTx *flow.Transaction) error {
	return nil
}

func (m *mockImpl) SignTransaction(ctx context.Context, tx *flow.Transaction, signerFlowAddress flow.Address, signer crypto.Signer) (*flow.Transaction, error) {
	return nil, nil
}

func (m *mockImpl) SignTransactionStr(ctx context.Context, txStr string, signerFlowAddress flow.Address, signer crypto.Signer) (*flow.Transaction, error) {
	return nil, nil
}

func (m *mockImpl) WaitTransactionSeal(ctx context.Context, id flow.Identifier) (*flow.TransactionResult, error) {
	return nil, nil
}

func (m *mockImpl) GetLatestBlock(ctx context.Context, isSealed bool, opts ...grpc.CallOption) (*flow.Block, error) {
	return nil, nil
}

func (m *mockImpl) Close() error {
	return nil
}
