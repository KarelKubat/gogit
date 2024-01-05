foo:
	@echo 'you can only "make install" here'
	@exit 1

install:
	go install gogit.go
