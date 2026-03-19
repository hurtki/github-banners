#!/bin/sh

ROOT_DIR=$(pwd)
RESULTS="/tmp/cov_$$.txt"
rm -f "$RESULTS"

run_service() {
    service=$1
    dir="${ROOT_DIR}/${service}"
    printf "\n%s\n" "$service"

    if [ ! -d "$dir" ]; then
        printf "  not found\n"
        echo "$service 0.0" >>"$RESULTS"
        return
    fi

    cd "$dir"
    total=0
    count=0
    tmp="/tmp/gt_$$.txt"

    go test -count=1 -cover ./... 2>&1 |
        grep -Ev 'no test files|skipped' >"$tmp"

    while IFS= read -r line; do
        [ -z "$line" ] && continue
        status=$(printf '%s' "$line" | awk '{print $1}')
        pkg=$(printf '%s' "$line" | awk '{print $2}' | awk -F'/' '{print $NF}')
        elapsed=$(printf '%s' "$line" | grep -o '[0-9]*\.[0-9]*s' | head -1)
        cov=$(printf '%s' "$line" | grep -o 'coverage: [0-9]*\.[0-9]*' | awk '{print $2}')
        [ -z "$elapsed" ] && elapsed="-"
        [ -z "$cov" ] && cov="0.0"

        if [ "$status" = "ok" ]; then
            printf "  ok   %-30s %6s  %s%%\n" "$pkg" "$elapsed" "$cov"
            total=$(awk -v a="$total" -v b="$cov" 'BEGIN{printf "%.4f",a+b}')
            count=$((count + 1))
        elif [ "$status" = "FAIL" ]; then
            printf "  FAIL %-30s\n" "$pkg"
            count=$((count + 1))
        fi
    done <"$tmp"
    rm -f "$tmp"

    avg="0.0"
    [ "$count" -gt 0 ] && avg=$(awk -v t="$total" -v c="$count" 'BEGIN{printf "%.1f",t/c}')
    printf "  avg coverage: %s%%\n" "$avg"

    echo "$service $avg" >>"$RESULTS"
    cd "$ROOT_DIR"
}

run_service "api"
run_service "renderer"
run_service "storage"

printf "\ncoverage summary\n"

while IFS= read -r line; do
    svc=$(printf '%s' "$line" | awk '{print $1}')
    val=$(printf '%s' "$line" | awk '{print $2}')
    printf "  %-10s %s%%\n" "$svc" "$val"
done <"$RESULTS"

overall=$(awk '{s+=$2; c++} END{printf "%.1f",s/c}' "$RESULTS")
printf "  %-10s %s%%\n" "total" "$overall"

rm -f "$RESULTS"
