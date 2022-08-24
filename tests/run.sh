#!/bin/bash

checkInterpreter() {
	[ -x "$1" ] && [ -f "$1" ] && interpreter=("$@")
}

interpreter=(go run .)
checkInterpreter ./direct
checkInterpreter ./whitespace

runWS() {
	"${interpreter[@]}" "$1"
}

fail() {
	local fmt="%s"
	if [ $# -gt 2 ]; then
		fmt=$1
		shift
	fi
	printf >&2 "Error: $fmt\n" "$@"
	exit 1
}

parseTest() {
	testInput=
	testExpect=

	if ! exec 8< "$1"
	then
		testState=ERROR
		testOutput="Error: unable to open test file"
		return 1
	fi

	local line
	while IFS= read -r -u 8 line
	do
		[[ "$line" = "input:"* ]] && testInput+="${line:6}"$'\n'
		[[ "$line" = "output:"* ]] && testExpect+="${line:7}"$'\n'
	done
	[[ "$line" = "input:"* ]] && testInput+="${line:6}"$'\n'
	[[ "$line" = "output:"* ]] && testExpect+="${line:7}"$'\n'

	exec 8<&-

	if [ "$testExpect" = "" ]; then
		testState=ERROR
		testOutput="Error: spec for expected output not found in test"
		return 1
	fi

	testInput=${testInput%$'\n'}
	testExpect=${testExpect%$'\n'}
	return 0
}

runTest() {
	local exitCode=0

	testState=RUN
	testOutput=$(runWS "$1" 2>&1 <<<"$testInput") exitCode=$?

	if [ $exitCode -eq 0 ] && [ "$testOutput" = "$testExpect" ]; then
		testState=OK
	else
		testState=FAIL

		[ $exitCode -eq 0 ] || testOutput+=$'\n'"Exit code: $exitCode"
	fi
}

handleTest() {
	testState=INIT
	testOutput=

	local name=${1##*/}
	printf "Test %s ... " "$name"

	parseTest "$1" && runTest "$1"

	printf "%s\n" "$testState"

	counts[$testState]=$(( ${counts[$testState]} +1 ))
	case "$testState" in
		OK) ;;
		ERROR) printf "\t%s\n" "$testOutput" ;;
		FAIL) showOutputs ;;
		*) showOutputs ;;
	esac
}

showOutputs() {
	if useBlock "$testOutput" || useBlock "$testExpect"
	then
		local out=${testExpect//$'\n'/$'\n\t\t'}
		printf "\tExpected output:\n\t\t%s\n" "$out"

		out=${testOutput//$'\n'/$'\n\t\t'}
		printf "\tActual output:\n\t\t%s\n" "$out"
	else
		printf "\tExpected output: %s\n" "$testExpect"
		printf "\tActual output  : %s\n" "$testOutput"
	fi
}

useBlock() {
	[ ${#1} -gt 40 ] || [[ "$1" = *$'\n'* ]]
}

declare -A counts=()

testDir=$(cd "${BASH_SOURCE[0]%/*}" && pwd) \
	|| fail "unable to enter tests dir"

printf "Using interpreter: %s\n" "${interpreter[*]}"

for testFile in "$testDir"/*.ws ; do
	handleTest "$testFile"
done

sep=
for s in "${!counts[@]}" ; do
	printf "%s%s: %s" "$sep" "$s" "${counts[$s]}"
	sep=", "
done
printf "\n"

[ ${#counts[@]} -eq 1 ] && [ "${counts["OK"]}" != "" ]
