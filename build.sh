#!/bin/bash

rm -Rf pkg/dashboard/frontend
cd node-dashboard
npm install
npm run build
cp -R build ../pkg/dashboard/frontend
