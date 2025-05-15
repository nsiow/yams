# Deployment

Although **yams** is meant to be run as a server, the project provides no strong opinions about how
or where that server is run. Infrastructure deployments vary greatly between companies, and very
few assumptions hold true.

The Go binary comprising the server can be easily built for cross-platform use, and the only real
requirement is the ability to serve or front an HTTP service on an open port.

If you have requests or ideas for deployment patterns to support more natively, please raise an
issue on GitHub.

# At a Glance

An overview for a recommended, basic **yams** deployment would be:

- Use [yams dump](./generating_data.md) commands to populate data sources in S3
  - Schedule these using whatever your friendly, neighborhood cron equivalent
- Deploy on a single (large) EC2 instance or container
- Point **yams** server at the cached data in S3

# Performance

The following tips should help in maximizing the performance of a **yams** deployment for
environments where the extra edge is needed:

- Use CPU-optimized EC2 instances: access simulation is almost an entirely CPU-bound endeavor and
  **yams** is heavily parallelized: use as many
- Split [data sources](./data_sources.md) into multiple files: while data loading will rarely be the
  limiting factor, it can become clunky as your sources approach GB scale. Individual data sources
  are loaded on their own goroutines, and so benefit from being partitioned by environment, resource
  type, etc
- Consider horizontal scaling: adding multiple instances of the **yams** server is an effective way
  to scale up. However, since every server independently refreshes its data sources it can
  introduce consistency issues if servers become out of sync
