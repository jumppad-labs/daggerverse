test_amd64: build_example
	cp ./example_app/bin/example_app_amd64 ./example_package_contents/bin/example
	dagger download build --files $(shell pwd)/example_package_contents --architecture amd64 --package-name "example" --version "0.0.1" --maintainer "Nic Jackson" --description "blah"

test_arm64: build_example
	cp ./example_app/bin/example_app_arm64 ./example_package_contents/bin/example
	dagger download build --files $(shell pwd)/example_package_contents --architecture arm64 --package-name "example" --version "0.0.1" --maintainer "Nic Jackson" --description "blah"

build_example:
	cd ./example_app && $(MAKE) build