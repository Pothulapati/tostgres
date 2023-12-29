# tostgres

Managed PostgreSQL offering built in a hour to demo how Temporal can be used for Infra Management. This
demo has a workflow called the `CreateTostgres` that allows you to create a new PostgreSQL instance and
add the relevant DNS record to get a `<xyz>.tostgres.cloud` domain.

## Running


Import your DigitalOcean API key as an environment variable:

```bash
export DO_TOKEN=<your token>
```

Then run the following command to start the Temporal locally

```bash
temporal server start-dev
```

Then run the following command to start the worker

```bash
go run ./...
```

Then run the following command to start the workflow

```bash
temporal workflow execute --workflow-id=test-2 --type=CreateTostgres --task-queue=default --input='{"name": "ptr", "region": "sfo3"}'
```

Once the workflow completes, You should be able to talk to the PostgreSQL instance using the following command:

```bash
psql -U ptr -d ptr -h ptr.tostgres.cloud  
```
