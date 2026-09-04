# AWS and Azure personas

This patch extends Chimera's built-in persona model with AWS EC2/EBS and Azure ARM Compute/Managed Disks.

## AWS persona

Set:

```bash
export CHIMERA_PERSONA=aws
export CHIMERA_USERNAME=AKIDCHIMERA
export CHIMERA_PASSWORD=chimera-secret
```

`CHIMERA_USERNAME` is the simulated AWS access-key ID and `CHIMERA_PASSWORD` is the simulated secret access key. AWS requests are authenticated with AWS Signature Version 4; the server verifies the actual HMAC signature, credential scope, service and payload hash.

Implemented EC2 Query actions:

- `DescribeInstances`
- `DescribeVolumes`
- `StartInstances`
- `StopInstances`
- `CreateSnapshot`
- `DescribeSnapshots`

Implemented EBS direct snapshot data plane:

- `GET /snapshots/{snapshotId}/blocks`
- `GET /snapshots/{snapshotId}/blocks/{blockIndex}?blockToken=...`

The block response is deterministic and carries checksum/data-length headers so migration readers can validate/resume reads.

Example with AWS CLI:

```bash
export AWS_ACCESS_KEY_ID=AKIDCHIMERA
export AWS_SECRET_ACCESS_KEY=chimera-secret
export AWS_DEFAULT_REGION=us-east-1
aws ec2 describe-instances --endpoint-url http://127.0.0.1:8989
aws ec2 create-snapshot --volume-id vol-chimera0001 --endpoint-url http://127.0.0.1:8989
```

## Azure persona

Set:

```bash
export CHIMERA_PERSONA=azure
export CHIMERA_USERNAME=11111111-2222-3333-4444-555555555555
export CHIMERA_PASSWORD=chimera-azure-token
```

For Azure, `CHIMERA_USERNAME` is the simulated subscription ID and `CHIMERA_PASSWORD` is the bearer token expected in `Authorization: Bearer ...`.

Implemented ARM surfaces:

- subscription and resource-group VM listing
- VM get
- VM instance view
- `start`
- `powerOff`
- `deallocate`
- `restart`
- managed disk get
- `beginGetAccess`
- Azure-style asynchronous operation polling
- SAS-style managed-disk byte access with HTTP Range support

The simulator accepts an `api-version` query parameter but intentionally does not bind behavior to one exact API version, allowing migration adapters using adjacent ARM versions to be tested.

## End-to-end coverage

`integration/cloud_personas_e2e_test.go` exercises:

### AWS

SigV4 -> DescribeInstances -> StartInstances -> CreateSnapshot -> ListSnapshotBlocks -> GetSnapshotBlock.

### Azure

Bearer auth -> VM list -> Start -> async operation poll -> InstanceView -> beginGetAccess -> async operation poll -> ranged disk download.
