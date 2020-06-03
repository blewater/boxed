# _remote_ workflows

Run your distributed tasks fluently over gRPC asynchronous streaming goodness.

## Why?
Because running your asynchronous distributed activities as a cohesive linear workflow is hard.
 
## Use Cases
* Distributed Docker deployments.

* Remote Batch Processing.

* Remote Service Provisioning.

* Data provenance and analysis.

* Any loosely-coupled server to remote activities.

## Examples
* _~~Guess~~_[Hack the shared (Diffie-Hellman) secret.](https://github.com/tradeline-tech/workflow/tree/development/examples/secret)

* Remote Tic-tac-toe play.

* Remote docker deployment.

## Features
* Dynamic task configuration bootstrapping.

* Monitor Your remote task hiccups centrally.

* Resume your workflow activities instantly, where you left off.

* Run your distributed tasks linearly as you declare them.

* Instant usage that scales to your workflow needs.

* _where the rubber hit the road_, tech borne-out from demanding distributed deployment conditions.

## Steps to create a workflow
1. Declare your tasks _sequentially_.

2. Start the gRPC server,

    Optionally, declare a Mongo connection.

3. Start your remote client to **run** your workflow :)

## Installation

`go get github.com/tradeline-tech/workflow`

## An inside look at the mechanics

![PlantUML model](http://www.plantuml.com/plantuml/svg/fLD1ReCm4Bpx5HjESA0FYA8g5SSgKGIfUgXwS613KS0RsIOflw-90m4ARQhwPArtPZopzb9fBdLPvEmZIn3sH7f7dupnM9E440lI-A9GiXsL8k6o0iSM7OP2Pxg22EK9vIl9mpwdCuifpp7M6Ga5ZZs3vb2zlJiiuPhlk49msZ94cck4JC2AH4eEOpTX0Dz_78Z07D9m4ypd4OgaAQvvWSzOkHuRD2_yqOiOaZMUcueBvpuFU0pCUZ9MJbpZq2QODQ8pQMaW5fFOPobuOpp6xdU_OIbMN9ZiI5PRhWxA7SFQi6nu1XHObONVYSlM3Fe2xrcqkARUq7G9OnAgBF1w_M7d3uE2Mhg-Tq35CSVwUMp9zbwDZ0SnbeFhtKxWaiKaWUbhdcllpTXQc-EqOLaAkpSetOxFkz_vK6uZAPMerDzS1tIi_atIHXTYFSJVexEAev_jDiQxmfZDUYn1JWexG8cwb8AnEuqqsuU8dmpD16pwBngAAv9rr9SeahB8lm00)
