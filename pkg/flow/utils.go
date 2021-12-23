package flow

import (
	"context"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flow-go-sdk/crypto/cloudkms"
)

func InitMemSigner(conf *InMemSignerConfig) (crypto.Signer, error) {
	serviceAccountSigAlgo := crypto.StringToSignatureAlgorithm(conf.SignatureAlgorithm)
	serviceAccountPrivateKey, err := crypto.DecodePrivateKeyHex(serviceAccountSigAlgo, conf.PrivateKeyHex)
	if err != nil {
		return nil, err
	}

	return crypto.NewInMemorySigner(serviceAccountPrivateKey, crypto.StringToHashAlgorithm(conf.PrivateKeyHashAlgoName)), nil
}

func InitKMSSigner(ctx context.Context, conf *KMSConfig, signerFlowAddress flow.Address) (crypto.Signer, error) {
	// init  kms client
	accountKMSKey := cloudkms.Key{
		ProjectID:  conf.KMSProjectID,
		LocationID: conf.KMSLocationID,
		KeyRingID:  conf.KMSKeyRingID,
		KeyID:      conf.KMSKeyID,
		KeyVersion: conf.KMSKeyVersion,
	}

	kmsClient, err := cloudkms.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	//addr := flow.HexToAddress(conf.SignerFlowAddress)

	return kmsClient.SignerForKey(
		context.Background(),
		signerFlowAddress,
		accountKMSKey,
	)
}
