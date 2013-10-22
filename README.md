# Snowflake

Snowflake is a HTTP server for generating unique ID numbers.

This is a Golang implementation of Twitter's Snowflake.

## Performance

TODO

## Spec

ID number is composed of:
- time - 41 bits (millisecond precision w/ a custom epoch gives us 69 years)
- configured machine id - 10 bits - gives us up to 1024 machines
- sequence number - 12 bits - rolls over every 4096 per machine (with protection to avoid rollover in the same ms)

This is equal to Twitter's Snowflake.

## Lisence

Apache License 2.0

## Links

- https://github.com/twitter/snowflake
