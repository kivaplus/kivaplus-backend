build:
	@cd lambda && \
		GOOS=linux GOARCH=amd64 go build -o bootstrap && \
		zip auth.zip bootstrap && \
		mv auth.zip ../dist && \
		rm bootstrap
clean:
	@cd dist && \
		rm auth.zip
