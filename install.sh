#!/bin/bash

app="httpserver"
path="/usr/local/"$app
echo $app, $path

go build $app
echo "======================================"
mkdir -p $path/ca
rm -rf $path/$app
rm -rf $path/$app.json
rm -rf $path/ca/*
cp $app $path
cp $app.json $path
cp ca/* $path/ca
$path/$app -d

sleep 1s
echo "ps -ef | grep -v grep | grep $app"
ps -ef | grep -v grep | grep $app
