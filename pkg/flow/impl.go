package flow

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/crypto"
	"google.golang.org/grpc"
)

type Flow struct {
	signerAddress flow.Address

	flowClient *client.Client
	signer     crypto.Signer
}

func newFlowProvider(c *Config) (*Flow, error) {
	if c.FlowEndPoint == "" {
		return nil, errors.New("flow access node URL is required")
	}
	flowClient, err := client.New(c.FlowEndPoint, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return &Flow{flowClient: flowClient}, nil
}

func (m *Flow) GetFlowClient() *client.Client {
	return m.flowClient
}

func (m *Flow) SignTransaction(ctx context.Context, tx *flow.Transaction, signerFlowAddress flow.Address, signer crypto.Signer) (*flow.Transaction, error) {
	account, err := m.flowClient.GetAccount(ctx, signerFlowAddress)
	if err != nil {
		return nil, fmt.Errorf("error getting account from flow: %w", err)
	}
	if err := tx.SignEnvelope(account.Address, account.Keys[0].Index, signer); err != nil {
		return nil, fmt.Errorf("error signing envelope with wallet: %w", err)
	}
	return tx, nil
}

func (m *Flow) SignTransactionStr(ctx context.Context, txStr string, signerFlowAddress flow.Address, signer crypto.Signer) (*flow.Transaction, error) {
	account, err := m.flowClient.GetAccount(ctx, signerFlowAddress)
	if err != nil {
		return nil, fmt.Errorf("error getting account from flow: %w", err)
	}

	latestBlock, err := m.flowClient.GetLatestBlock(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("error getting latest block from flow: %w", err)
	}
	tx := flow.NewTransaction().
		SetScript([]byte(txStr)).SetGasLimit(100).
		SetProposalKey(account.Address, account.Keys[0].Index, account.Keys[0].SequenceNumber).
		SetReferenceBlockID(latestBlock.ID).
		SetPayer(account.Address).
		AddAuthorizer(account.Address)
	return m.SignTransaction(ctx, tx, signerFlowAddress, signer)
}

func (m *Flow) SendSignedTransaction(ctx context.Context, signedTx *flow.Transaction) error {
	fmt.Printf("running tx id %s with seq %d\n", signedTx.ID().Hex(), signedTx.ProposalKey.SequenceNumber)
	if err := m.flowClient.SendTransaction(context.Background(), *signedTx); err != nil {
		return fmt.Errorf("error sending tx: %w", err)
	}
	return nil
}

func (m *Flow) WaitTransactionSeal(ctx context.Context, id flow.Identifier) (*flow.TransactionResult, error) {
	for ctx.Err() == nil {
		result, err := m.flowClient.GetTransactionResult(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("error getting transaction result: %w", err)
		}
		if result.Status == flow.TransactionStatusSealed {
			fmt.Printf("Transaction %x sealed\n", id)
			return result, result.Error
		}

		if deadline, hasDeadline := ctx.Deadline(); hasDeadline && deadline.Before(time.Now()) {
			return nil, fmt.Errorf("error getting transaction result within timeout")
		}
		time.Sleep(time.Duration(100) * time.Millisecond)
	}
	return nil, ctx.Err()
}

func (m *Flow) GetLatestBlock(ctx context.Context, isSealed bool, opts ...grpc.CallOption) (*flow.Block, error) {
	return m.flowClient.GetLatestBlock(ctx, isSealed, opts...)
}

func (m *Flow) Close() error {
	return m.flowClient.Close()
}
