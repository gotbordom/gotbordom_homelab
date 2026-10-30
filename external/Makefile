.POSIX:

default: apply

terraform.tfvars:
	cp terraform.tfvars.example ${@}
	nvim ${@}

.terraform.lock.hcl: versions.tf terraform.tfvars
	tofu init
	touch ${@}

namespaces:
	ansible-playbook namespaces.yml

plan: .terraform.lock.hcl
	tofu plan

apply: .terraform.lock.hcl namespaces
	tofu apply -auto-approve
