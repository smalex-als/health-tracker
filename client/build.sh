#!/bin/bash

mvn clean install
rm -rf ../server/app/static/js/*
cp target/tracker-1.0/tracker/* ../server/app/static/js/

