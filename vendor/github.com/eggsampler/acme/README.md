# eggsampler/acme

[![GoDoc](https://godoc.org/github.com/eggsampler/acme?status.svg)](https://godoc.org/github.com/eggsampler/acme)

## About

`eggsampler/acme` is a Go client library implementation for [ACME v2 revision 10](https://tools.ietf.org/html/draft-ietf-acme-acme-10), specifically for use with the [Let's Encrypt](https://letsencrypt.org/) service. 

The library is designed to provide a wrapper over exposed directory endpoints and provide objects in easy to use structures.

## Example

A simple [certbot](https://certbot.eff.org/)-like example is provided in the examples/certbot directory. This code demonstrates account registation, new order submission, fulfilling challenges, finalising an order and fetching the issued certificate chain.

An example of how to use the autocert package is also provided in examples/autocert.

## Tests

### Boulder

Tests can be run against a local instance of [boulder](https://github.com/letsencrypt/boulder) running the `config-next` configuration.

This needs to have the `chaltestsrv` responding to http01 challenges. This is currently disabled by default and can be enabled by editing `test/startservers.py` and ensure `chaltestsrv` is running with the flag `--http01 :5002` instead of `--http01 ""`

By default, tests run using the boulder client unless the environment variable `ACME_SERVER` is set to `pebble`, eg: `ACME_SERVER=boulder go test -v`

### Pebble

Alternatively, tests can be run against a local instance of [pebble](https://github.com/letsencrypt/pebble) running with the `PEBBLE_VA_ALWAYS_VALID=1` environment variable.

Currently pebble does not support account key changes or deactivating authorizations, so these tests are disabled.

To run tests using the pebble client the environment variable `ACME_SERVER` must be set to `pebble`, eg: `ACME_SERVER=pebble go test -v`