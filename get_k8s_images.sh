#!/bin/bash

kubectl get pods --all-namespaces -o jsonpath="{.items[*].spec.containers[*].image}"
