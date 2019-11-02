# openstack-quota-collector

## Run Openstack Quota Collector as Docker container

1. Build the docker container.

```bash
docker build -t os-quota-collector:latest .
```

2. Fill in the .env file with the relevant credentials, urls and tenant information (these are the variables you find in the Openstack RC File v2 or v3).
3. Run the docker container, passing it your .env file.

```bash
docker run -p 9080:9080 --env-file .env os-quota-collector:latest
```

4. Visit `localhost:9080` to view the metrics.

