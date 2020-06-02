# _remote_ workflows

Run your distributed tasks fluently to **_run_** over gRPC asynchronous streaming goodness.
 
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

## Internals
![PlantUML model](http://www.plantuml.com/plantuml/svg/fLF1hjem4BpxAvQSaWFz05HLL3bMHIYXwg7gmIGBZKYyo7QW-VjkC2Oak3oU7WSKhNTcTcRjfR5IsxQfnP-gHDWHsItz5K8MbZHas9357mQrs5AhIgaxY5mn4gXiPJl8KPzG_oHzsDLq5WNLDHQs8cKWScVW3ysltv4efPLRQH2YMnboAotoFsdc5OCgm7p-PQ802zXdwtIno56LlhYI3Nx3Bax4fFdpF3W5awpmx2indCSya0ZWtrVsSP8Mfnxxiu3En4yqqIX8xQez924uktoqcRNgk-JGBUbk8yt1n26ioyXbng3KgV0yGqgknkWBVEJCuNpTqzmJyjYcRV0w_hJpoNwWb-xdJgavrjXVYGsvtgkHS12X7E9aE85B5lybu3v_HKTNnTLyZayCwrQvRqRPs_gIrvrFvTQQzXgDUP_hb8xTHqc6w3Bib_XsT4un-CQ6DRuKncmsOabhCdW7r6XRhOHuqOZ__9JB4zE4OOzFa95NWBWJNV8yRLF_0m00)
