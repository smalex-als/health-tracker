#!/bin/bash

dev_appserver.py \
  --search_indexes_path ../data/index \
  --datastore_path ../data/db \
  --host 192.168.1.203 \
  --admin_host 192.168.1.203 \
  .
