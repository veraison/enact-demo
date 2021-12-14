CMD := corim-repackage
SRC := $(wildcard *.go)

all: $(CMD); ./$(CMD)

$(CMD): $(SRC); go build .

clean: ; $(RM) $(CMD)