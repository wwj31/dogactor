#!/bin/bash

if [ "$1" == "f" ]; then
  echo "format proto..."
  for file in `ls ./*.proto`
  do
      clang-format -i -style="{AlignConsecutiveAssignments: true,AlignConsecutiveDeclarations: true,AllowShortFunctionsOnASingleLine: None,BreakBeforeBraces: GNU,ColumnLimit: 0,IndentWidth: 4,Language: Proto}" $file
  done
fi

binpath=../../../../../protoc/mac
#$binpath/protoc --plugin protoc-gen-go=$binpath/protoc-gen-go-mac -I=./ --go_out=./ ./*.proto

$binpath/protoc --plugin protoc-gen-gofast=$binpath/protoc-gen-gofast -I=./ --gofast_out=plugins=grpc:. *.proto