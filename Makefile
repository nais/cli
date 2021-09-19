liberator = $(wildcard ../liberator)

.PHONY: nais-cli liberator

local:
	go build -o tool/nais ./main/nais_cli/

nais-cli: liberator
	go install main/nais_cli/nais_cli.go

test:
	go test ./... -count=1 -coverprofile cover.out -short

liberator:
ifneq ($(wildcard $(liberator)),)
	@echo "Found $(liberator)."
	@echo "Use cmd: kubectl apply -f ../liberator/config/crd/bases"
	@echo "Add CRDs to your minikube local setup"
else
	@echo "Did not find $(liberator)."
	$(error please clone: https://github.com/nais/liberator)
endif
