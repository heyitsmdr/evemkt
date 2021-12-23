.PHONY: publish
publish:
	docker build -t us-east4-docker.pkg.dev/evemkt/evemkt/evemkt-server:latest .
	docker push us-east4-docker.pkg.dev/evemkt/evemkt/evemkt-server:latest