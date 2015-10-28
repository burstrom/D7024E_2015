#!/bin/bash
docker run  -d -i -t -p $2:$2 --net=host --name $1 burstrom/reposky localhost:$2