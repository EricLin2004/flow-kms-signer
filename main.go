package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	localFlow "flow-kms-signer/pkg/flow"
	txUtils "flow-kms-signer/pkg/templates"

	"github.com/onflow/flow-go-sdk/crypto"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto/cloudkms"
	"github.com/urfave/cli/v2"
)

type kmsConf struct {
	projectID     string
	keyRing       string
	keyVersion    string
	key           string
	signerAddress string
}

type AccountInfo struct {
	Account    *flow.Account
	AccountKey *flow.AccountKey
	Signer     crypto.Signer
}

func main() {
	app := &cli.App{}
	app.UseShortOptionHandling = true
	app.Commands = []*cli.Command{
		{
			Name:  "sasm",
			Usage: "sign and send multiple cadence transactions using different arguments with the corresponding KMS key and send transaction to access node",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "project", Aliases: []string{"p"}, EnvVars: []string{"KMS_PROJECT"}, Required: true},
				&cli.StringFlag{Name: "keyring", Aliases: []string{"kr"}, EnvVars: []string{"KMS_KEYRING"}, Required: true},
				&cli.StringFlag{Name: "keyversion", Aliases: []string{"kv"}, EnvVars: []string{"KMS_KEYVERSION"}, Required: true},
				&cli.StringFlag{Name: "key", Aliases: []string{"k"}, EnvVars: []string{"KMS_KEY"}, Required: true},
				&cli.StringFlag{Name: "signeraddress", Aliases: []string{"s"}, EnvVars: []string{"SIGNER_ADDRESS"}, Required: true},
				&cli.StringFlag{Name: "flowaccessnode", Aliases: []string{"f"}, EnvVars: []string{"FLOW_ACCESS_NODE"}, Required: true},
				&cli.StringFlag{Name: "cadencefilepath", Aliases: []string{"c"}, Required: true},
				&cli.StringFlag{Name: "cadencearguments", Aliases: []string{"ca"}},
			},
			Action: kmsSigner,
		}
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func kmsSigner(c *cli.Context) error {
	ctx := context.Background()
	projectID := c.String("project")
	keyRing := c.String("keyring")
	keyVersion := c.String("keyversion")
	key := c.String("key")
	signerAddress := c.String("signeraddress")
	flowAccessNode := c.String("flowaccessnode")
	cadenceFilePath := c.String("cadencefilepath")

	if flowAccessNode == "" || projectID == "" || keyRing == "" || keyVersion == "" || key == "" || signerAddress == "" || cadenceFilePath == "" {
		return errors.New("missing arguments in sign-and-send command")
	}

	// KMS Configuration
	kmsConfig := kmsConf{
		projectID:     projectID,
		keyRing:       keyRing,
		keyVersion:    keyVersion,
		key:           key,
		signerAddress: signerAddress,
	}

	// Flow client
	flowProviderConfig := &localFlow.Config{
		FlowEndPoint:    flowAccessNode,
		MockFlowEnabled: false,
	}

	flowProvider, err := localFlow.New(flowProviderConfig)
	if err != nil {
		panic(fmt.Errorf("failed to deploy connect to flow access node: %w", err))
	}

	// Get Flow account details, key index, sequence number, also KMS signer
	accountInfo := getAccountInfo(ctx, flowProvider, kmsConfig)
	accountAddress := accountInfo.Account.Address
	accountKey := accountInfo.AccountKey

	// Setup Cadence script with arguments
	// Cadence arguments mapping
	cadenceArguments := c.String("cadencearguments")
	if cadenceArguments != "" {
		cadenceArgsSplit := strings.Split(cadenceArguments, ";")

		for i := 0; i < len(cadenceArgsSplit); i++ {
			cadenceArgumentsMap := make(map[string]string)
			cadenceArgsPerTx := strings.Split(cadenceArgsSplit[i], ",")

			for j := 0; j < len(cadenceArgsPerTx); j++ {
				cadenceArgumentsMap[fmt.Sprintf("Arg%d", i)] = cadenceArgsPerTx[j]

				txScript := txUtils.ParseCadenceTemplateV2(cadenceFilePath, cadenceArgumentsMap)

				latestBlock, _ := flowProvider.GetLatestBlock(context.Background(), true)

				setupTx :=
					flow.NewTransaction().
						SetScript(txScript).
						SetGasLimit(100).
						SetProposalKey(
							accountAddress,
							accountKey.Index,
							accountKey.SequenceNumber).
						SetReferenceBlockID(latestBlock.ID).
						SetPayer(accountAddress).
						AddAuthorizer(accountAddress)

				// Sign and send to Flow access node
				signedTx, err := flowProvider.SignTransaction(ctx, setupTx, accountAddress, accountInfo.Signer)
				if err != nil {
					return err
				}

				// Wait for seal
				result, err := SendSignedTransactionAndWaitForSeal(ctx, flowProvider, signedTx)
				if err != nil {
					return err
				}

				fmt.Println("==> Transaction signed and sent")
				fmt.Printf("Status: %s\n", result.Status)
				fmt.Printf("Events: %s\n", result.Events)

				accountKey.SequenceNumber++
			}
		}
	}

	return nil
}

func SendSignedTransactionAndWaitForSeal(ctx context.Context, flowProvider localFlow.Provider, tx *flow.Transaction) (*flow.TransactionResult, error) {
	if err := flowProvider.SendSignedTransaction(ctx, tx); err != nil {
		return nil, err
	}
	return flowProvider.WaitTransactionSeal(ctx, tx.ID())
}

func getAccountInfo(ctx context.Context, flowProvider localFlow.Provider, kmsConfig kmsConf) *AccountInfo {
	accountKMSKey := cloudkms.Key{
		ProjectID:  kmsConfig.projectID,
		LocationID: "global",
		KeyRingID:  kmsConfig.keyRing,
		KeyID:      kmsConfig.key,
		KeyVersion: kmsConfig.keyVersion,
	}

	kmsClient, err := cloudkms.NewClient(ctx)
	if err != nil {
		panic(fmt.Errorf("err create new cloud kms client %s", err))
	}

	addr := flow.HexToAddress(kmsConfig.signerAddress)
	accountKMSSigner, err := kmsClient.SignerForKey(
		ctx,
		addr,
		accountKMSKey,
	)
	if err != nil {
		panic(fmt.Errorf("err create kms signer: %s", err))
	}

	account, err := flowProvider.GetFlowClient().GetAccount(ctx, addr)
	if err != nil {
		panic(fmt.Errorf("failed to get account: %s", err))
	}

	return &AccountInfo{
		Account:    account,
		AccountKey: account.Keys[0],
		Signer:     accountKMSSigner,
	}
}
