# Usage

    curl 'https://results.zone/grom-relay-osen-2023/races/6991/results.json?q%5Bteam_eq%5D=VK'  | jq '.' > ~/data/grom/grom-relay-osen-2023/results.json
    go run cmd/parse_results/main.go --input-file ~/data/grom/grom-relay-osen-2023/results.json
