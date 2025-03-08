#DevOps Makefile

NAME := gotest
OBJS = $(NAME) main.go
SRCPATH= /data/GoProject/$(NAME)
TARGET = /data/Buildfile/

.PHONY: $(NAME)
$(NAME): $(OBJS)
	@cd $(SRCPATH) && \
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o $@ $^ && \
	mv $(NAME) $(TARGET) && \
	cd $(TARGET) && tar -zcf $(NAME).tgz *

clean:
	rm -f $(TARGET)*
