directory = $(wildcard ../liberator)

.PHONY: nais-d liberator

local:
	go build -o tool/nais-d main/nais-d/*.go

nais-d: liberator
	go install main/nais-d/nais-d.go

test:
	go test ./... -count=1 -coverprofile cover.out -short

liberator:
ifneq ($(wildcard $(directory)),)
	@echo "Found $(directory)."
	@echo "Use cmd: kubectl apply -f ../liberator/config/crd/bases"
	@echo "Add CRDs to your minikube local setup"
else
	@echo "Did not find $(directory)."
	$(error please clone: https://github.com/nais/liberator)
endif