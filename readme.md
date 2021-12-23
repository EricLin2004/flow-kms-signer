```
GO111MODULE=on go get github.com/urfave/cli/v2
```

Run `go install`

Main command is

`flow-kms-signer sasm`

Required flags and fields are in `.env.example`

```
KMS_PROJECT=
KMS_KEYRING=
KMS_KEYVERSION=
KMS_KEY=
SIGNER_ADDRESS=
FLOW_ACCESS_NODE=
```

Set these into env vars if you want to be signing with the same key, same flow address and send them all to same access node.

Your google auth login `gcloud auth application-default login` must have the access to the GCP Project and KMS credentials above, if not you will get a KMS error when trying to sign and send.

The cadence file to run will be provided in a flag
`flow-kms-signer sasm -c /path/to/cadence/file.cdc`

If you have arguments you want to provide for the cadence transaction, use the flag `-ca` with comma separated arguments. It will replace `{{..Arg0}}` to `{{.ArgN}}` in the cadence file itself.

For example

if `file.cdc` looks like

```
import FungibleToken from 0x{{.Arg0}}
import DapperUtilityCoin from 0x{{.Arg1}}

transaction {
    prepare(signer: AuthAccount) {}
}
```
Running `flow-kms-signer sasm -c /path/to/cadence/file.cdc -ca value1,two` would attempt to sign the following:
```
import FungibleToken from 0xvalue1
import DapperUtilityCoin from 0xtwo

transaction {
    prepare(signer: AuthAccount) {}
}
```

You can run multiple tx one after the other by using arguments separated by semi-colons. It will wait for sealing before sending the next transaction. For example: `flow-kms-signer sasm -c /path/to/cadence/file.cdc -ca value1,two;value2,three` will attempt to send two transactions using the same `file.cdc` template with arguments `value1,two` for the first and `value2,three` for the second.