# Remote Workflows

Execute your streaming workflow activities partitioned to server and remote tasks for repeatable, reliable asynchronous workflow processes.

## Features
* Bi-Directional tasks' configuration updates, powered by gRPC streaming.

* Monitor Your Workflow Hiccups.

* Resume your workflow activities after disuptions.

* Simple implementation that scales to your workflow needs.

## Steps to create a workflow
1. Declare your server and remote tasks execution order.

2. Add their shared configuration hooks.

3. Add your tasks' execution logic.

4. Optionally, add data-store state persistence.

5. Run your workflow :)

## Installation

go get github.com/tradeline-tech/workflow

## Use Cases
* Remote Batch Processing.

* Remote Service Provisioning.

* Data provenance and analysis.

* Any loosely-coupled server to remote activities.

## Examples
* Guess the shared (Diffie-Hellman) secret.

* Remote Tic-tac-toe play.

* Remote docker deployment.
