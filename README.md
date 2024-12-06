ClamAV API

Couldn't find a good microservice for clamAV in go so I created one.

DockerFile uses version 1.4 of clam to create the image, this includes all you need since its the official image

Uses tcp to communicate with clamAV using the documentation from https://linux.die.net/man/8/clamd

The api will wait to clamAV to be ready before starting
