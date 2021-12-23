package flow

import (
	"context"
	"errors"
	"io"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/crypto"
	"google.golang.org/grpc"
)

var NotImplementedError = errors.New("Not implemented")

type (
	Provider interface {
		io.Closer

		GetFlowClient() *client.Client
		GetLatestBlock(ctx context.Context, isSealed bool, opts ...grpc.CallOption) (*flow.Block, error)
		SendSignedTransaction(ctx context.Context, signedTx *flow.Transaction) error
		SignTransaction(ctx context.Context, tx *flow.Transaction, signerFlowAddress flow.Address, signer crypto.Signer) (*flow.Transaction, error)
		SignTransactionStr(ctx context.Context, tx string, signerFlowAddress flow.Address, signer crypto.Signer) (*flow.Transaction, error)
		WaitTransactionSeal(ctx context.Context, id flow.Identifier) (*flow.TransactionResult, error)
	}
)
