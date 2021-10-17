# _remote_ workflows

Run your distributed tasks fluently over gRPC asynchronous streaming goodness. :traffic_light:

![Workflow run gif](./workflow.gif)

## Why?
Because running asynchronous distributed activities as a cohesive linear workflow is hard.

## Features
* Define your _distributed_ tasks *sequentially* running 100% pure undomesticated Go.

* Live, cross-task declarative messaging, and configuration updates.

* Monitor your remote tasks' progress centrally.

* Because hiccups happen, resume your workflow activities where you left off.

* _where the rubber hit the road_, tech borne-out from demanding network deployment conditions.

## Use Cases
* Distributed Docker, K8s deployments.

* Remote Batch Processing.

* Remote Service Provisioning.

* Data provenance and analysis.

* Any loosely-coupled server to remote activities.

## Examples
* [Diffie-Hellman identification protocol with Knowledge-of-Exponent proof](https://github.com/blewater/boxed/tree/development/examples/secret)
```
For a generator g of group G; 

Prover knows exponent x in y=g^x
1. Verifier selects a randomly k value, and sends v=g^k to the prover.
2. Prover evaluates r=v^x and sends it to the verifier.
3. Verifier accepts iff r=yk.
```

* Distributed Containers deployment. (Not Implemented Yet).

* Remote Tic-tac-toe play. (Not Implemented Yet).

## Steps to create a workflow
1. Declare your tasks _sequentially_.

  _Optionally, declare a Mongo connection._

2. Start the gRPC server,

3. Start your remote client to **run** your workflow :)

## Installation

`go get github.com/tradeline-tech/workflow`

## An inside look at the mechanics

![PlantUML model](http://www.plantuml.com/plantuml/svg/fLD1ReCm4Bpx5HjESA0FYA8g5SSgKGIfUgXwS613KS0RsIOflw-90m4ARQhwPArtPZopzb9fBdLPvEmZIn3sH7f7dupnM9E440lI-A9GiXsL8k6o0iSM7OP2Pxg22EK9vIl9mpwdCuifpp7M6Ga5ZZs3vb2zlJiiuPhlk49msZ94cck4JC2AH4eEOpTX0Dz_78Z07D9m4ypd4OgaAQvvWSzOkHuRD2_yqOiOaZMUcueBvpuFU0pCUZ9MJbpZq2QODQ8pQMaW5fFOPobuOpp6xdU_OIbMN9ZiI5PRhWxA7SFQi6nu1XHObONVYSlM3Fe2xrcqkARUq7G9OnAgBF1w_M7d3uE2Mhg-Tq35CSVwUMp9zbwDZ0SnbeFhtKxWaiKaWUbhdcllpTXQc-EqOLaAkpSetOxFkz_vK6uZAPMerDzS1tIi_atIHXTYFSJVexEAev_jDiQxmfZDUYn1JWexG8cwb8AnEuqqsuU8dmpD16pwBngAAv9rr9SeahB8lm00)
