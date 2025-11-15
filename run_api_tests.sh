#!/usr/bin/env bash
# This script now supports proxy mode (default) using the NGINX reverse proxy at port 8080.
# Use `./run_api_tests.sh direct` to bypass the proxy and hit services individually.

set -u -o pipefail

GREEN="\033[32m"
RED="\033[31m"
YELLOW="\033[33m"
RESET="\033[0m"

TEST_PRODUCT_NAME="Test Product"
TEST_PRODUCT_PRICE="123.45"
UPDATED_PRODUCT_PRICE="543.21"

DEFAULT_MODE="proxy"
MODE="$DEFAULT_MODE"
TARGET="all"

BASE_GO_URL=""
BASE_NODE_URL=""
BASE_DOTNET_URL=""

SERVICES=(
  "go|Go service"
  "node|Node service"
  "dotnet|.NET service"
)

JQ_AVAILABLE=0
if command -v jq >/dev/null 2>&1; then
  JQ_AVAILABLE=1
fi

SERVICE_SUMMARY=()
TOTAL_SERVICES=0
FAILED_SERVICES=0
LAST_PRODUCT_ID=""

function print_ok() {
  printf "%b✔ %s%b\n" "$GREEN" "$1" "$RESET"
}

function print_fail() {
  printf "%b✘ %s%b\n" "$RED" "$1" "$RESET"
}

function print_info() {
  printf "%bℹ %s%b\n" "$YELLOW" "$1" "$RESET"
}

function usage() {
  cat <<'USAGE'
Usage: ./run_api_tests.sh [proxy|direct] [go|node|dotnet|all]
  Mode:   proxy (default) hits services via NGINX, direct bypasses the proxy.
  Target: defaults to "all". You can still run a single service (go|node|dotnet).
Examples:
  ./run_api_tests.sh                # proxy mode, all services
  ./run_api_tests.sh direct go      # direct mode, Go service only
  ./run_api_tests.sh node           # proxy mode, Node service only
USAGE
}

function parse_arguments() {
  MODE="$DEFAULT_MODE"
  TARGET="all"
  local mode_set=0
  local target_set=0

  local arg
  for arg in "$@"; do
    local normalized
    normalized=$(printf '%s' "$arg" | tr '[:upper:]' '[:lower:]')
    case "$normalized" in
      proxy|direct)
        if (( mode_set )); then
          usage
          exit 1
        fi
        MODE="$normalized"
        mode_set=1
        ;;
      go|node|dotnet|all)
        if (( target_set )); then
          usage
          exit 1
        fi
        TARGET="$normalized"
        target_set=1
        ;;
      *)
        usage
        exit 1
        ;;
    esac
  done
}

function configure_base_urls() {
  if [[ "$MODE" == "proxy" ]]; then
    BASE_GO_URL="http://localhost:8080/go"
    BASE_NODE_URL="http://localhost:8080/node"
    BASE_DOTNET_URL="http://localhost:8080/dotnet"
  else
    BASE_GO_URL="http://localhost:8081"
    BASE_NODE_URL="http://localhost:8082"
    BASE_DOTNET_URL="http://localhost:8083"
  fi

}

function service_base_url() {
  local key=$1
  case "$key" in
    go)
      printf '%s' "$BASE_GO_URL"
      ;;
    node)
      printf '%s' "$BASE_NODE_URL"
      ;;
    dotnet)
      printf '%s' "$BASE_DOTNET_URL"
      ;;
    *)
      printf ''
      ;;
  esac
}

function print_routing_info() {
  local mode_upper
  mode_upper=$(printf '%s' "$MODE" | tr '[:lower:]' '[:upper:]')
  if [[ "$MODE" == "proxy" ]]; then
    echo "Running tests in ${mode_upper} mode (NGINX @ :8080)"
  else
    echo "Running tests in ${mode_upper} mode (direct service ports)"
  fi
  echo "Routing:"
  echo "  Go   -> ${BASE_GO_URL}"
  echo "  Node -> ${BASE_NODE_URL}"
  echo "  .NET -> ${BASE_DOTNET_URL}"
}

function ensure_proxy_health() {
  if [[ "$MODE" != "proxy" ]]; then
    return 0
  fi

  local proxy_health_url="http://localhost:8080/health"
  print_info "Validating NGINX proxy health (${proxy_health_url})"

  local raw
  if ! raw=$(curl -s -S "$proxy_health_url" -w "HTTPSTATUS:%{http_code}"); then
    print_fail "curl ${proxy_health_url} failed"
    return 1
  fi
  local status="${raw##*HTTPSTATUS:}"
  if [[ "$status" == "200" ]]; then
    print_ok "NGINX proxy responded with 200"
    return 0
  fi

  print_fail "NGINX proxy health check expected 200 but got ${status}"
  return 1
}

function http_request() {
  local method=$1
  local url=$2
  local data=${3-}
  local response

  if [[ $# -ge 3 ]]; then
    response=$(curl -s -S -m 15 -X "$method" "$url" \
      -H "Content-Type: application/json" \
      -H "Accept: application/json" \
      -d "$data" -w "HTTPSTATUS:%{http_code}")
  else
    response=$(curl -s -S -m 15 -X "$method" "$url" \
      -H "Accept: application/json" \
      -w "HTTPSTATUS:%{http_code}")
  fi

  local curl_exit=$?
  if [[ $curl_exit -ne 0 ]]; then
    return $curl_exit
  fi

  printf "%s" "$response"
  return 0
}

HTTP_STATUS=""
HTTP_BODY=""

function parse_http_response() {
  local raw=$1
  HTTP_STATUS="${raw##*HTTPSTATUS:}"
  HTTP_BODY="${raw%HTTPSTATUS:*}"
}

function wait_for_service() {
  local name=$1
  local base_url=$2
  local retries=20
  local attempt=1
  local health_url="${base_url}/health"

  print_info "Waiting for ${name} to become ready (${health_url})"
  while (( attempt <= retries )); do
    local status
    status=$(curl -s -o /dev/null -w "%{http_code}" "$health_url" || true)
    if [[ "$status" == "200" ]]; then
      print_ok "[${name}] Service is reachable"
      return 0
    fi
    sleep 1
    ((attempt++))
  done

  print_fail "[${name}] Service did not become ready within 20 seconds"
  return 1
}

function validate_body_has_status_ok() {
  if [[ $JQ_AVAILABLE -eq 1 ]]; then
    printf '%s' "$HTTP_BODY" | jq -e '.status // "" | ascii_downcase == "ok"' >/dev/null 2>&1
    return $?
  fi

  [[ "$HTTP_BODY" == *'"status"'* && "$HTTP_BODY" == *'ok'* ]]
}

function validate_name_and_price() {
  local expected_name=$1
  local expected_price=$2

  if [[ $JQ_AVAILABLE -eq 1 ]]; then
    printf '%s' "$HTTP_BODY" | jq -e --arg name "$expected_name" --arg price "$expected_price" '
      (.name // "" == $name) and ((.price|tostring) == $price)
    ' >/dev/null 2>&1
    return $?
  fi

  [[ "$HTTP_BODY" == *"\"name\""*"$expected_name"* && "$HTTP_BODY" == *"\"price\""*"${expected_price}"* ]]
}

function extract_product_id() {
  if [[ $JQ_AVAILABLE -eq 1 ]]; then
    jq -r '(.id // .productId // .ID // .Id // ._id // empty)' <<<"$HTTP_BODY"
    return
  fi

  local regex='"(id|productId|ID|Id|_id)"[[:space:]]*:[[:space:]]*"?([A-Za-z0-9._-]+)"?'
  if [[ "$HTTP_BODY" =~ $regex ]]; then
    printf '%s' "${BASH_REMATCH[2]}"
    return
  fi

  printf ''
}

function test_health() {
  local name=$1
  local base_url=$2

  local raw
  if ! raw=$(http_request "GET" "${base_url}/health"); then
    print_fail "[${name}] /health request failed"
    return 1
  fi

  parse_http_response "$raw"

  if [[ "$HTTP_STATUS" != "200" ]]; then
    print_fail "[${name}] /health expected 200, got ${HTTP_STATUS}"
    return 1
  fi

  if ! validate_body_has_status_ok; then
    print_fail "[${name}] /health body did not include status=ok"
    return 1
  fi

  print_ok "[${name}] /health passed"
  return 0
}

function test_create_product() {
  local name=$1
  local base_url=$2

  local payload
  payload=$(cat <<'JSON'
{
  "name": "Test Product",
  "price": 123.45,
  "tags": ["test", "codex"]
}
JSON
)

  local raw
  if ! raw=$(http_request "POST" "${base_url}/products" "$payload"); then
    print_fail "[${name}] POST /products request failed"
    return 1
  fi

  parse_http_response "$raw"

  if [[ "$HTTP_STATUS" != "201" ]]; then
    print_fail "[${name}] POST /products expected 201, got ${HTTP_STATUS}"
    return 1
  fi

  if ! validate_name_and_price "$TEST_PRODUCT_NAME" "$TEST_PRODUCT_PRICE"; then
    print_fail "[${name}] POST /products response did not echo name/price"
    return 1
  fi

  local product_id
  product_id=$(extract_product_id)
  if [[ -z "$product_id" ]]; then
    print_fail "[${name}] POST /products missing product id"
    return 1
  fi

  LAST_PRODUCT_ID="$product_id"
  print_ok "[${name}] Created product (id=${product_id})"
  return 0
}

function test_get_product() {
  local name=$1
  local base_url=$2
  local product_id=$3
  local expected_name=$4
  local expected_price=$5
  local context=${6:-""}

  local raw
  if ! raw=$(http_request "GET" "${base_url}/products/${product_id}"); then
    print_fail "[${name}] GET /products/${product_id} request failed"
    return 1
  fi

  parse_http_response "$raw"

  if [[ "$HTTP_STATUS" != "200" ]]; then
    print_fail "[${name}] GET /products/${product_id} expected 200, got ${HTTP_STATUS}"
    return 1
  fi

  if ! validate_name_and_price "$expected_name" "$expected_price"; then
    print_fail "[${name}] GET /products/${product_id} validation failed"
    return 1
  fi

  local label="GET /products/${product_id}"
  if [[ -n "$context" ]]; then
    label="${label} (${context})"
  fi
  print_ok "[${name}] ${label} passed"
  return 0
}

function test_list_products() {
  local name=$1
  local base_url=$2

  local raw
  if ! raw=$(http_request "GET" "${base_url}/products"); then
    print_fail "[${name}] GET /products request failed"
    return 1
  fi

  parse_http_response "$raw"

  if [[ "$HTTP_STATUS" != "200" ]]; then
    print_fail "[${name}] GET /products expected 200, got ${HTTP_STATUS}"
    return 1
  fi

  if [[ $JQ_AVAILABLE -eq 1 ]]; then
    if ! printf '%s' "$HTTP_BODY" | jq -e 'type == "array" and length > 0' >/dev/null 2>&1; then
      print_fail "[${name}] GET /products did not return a non-empty array"
      return 1
    fi
  else
    if [[ "$HTTP_BODY" != *"["* ]] || [[ "$HTTP_BODY" != *"{"* ]]; then
      print_fail "[${name}] GET /products response not array-like"
      return 1
    fi
  fi

  print_ok "[${name}] GET /products listing passed"
  return 0
}

function test_update_product() {
  local name=$1
  local base_url=$2
  local product_id=$3

  local payload
  payload=$(cat <<JSON
{
  "name": "${TEST_PRODUCT_NAME}",
  "price": ${UPDATED_PRODUCT_PRICE},
  "tags": ["test", "codex", "updated"]
}
JSON
)

  local raw
  if ! raw=$(http_request "PUT" "${base_url}/products/${product_id}" "$payload"); then
    print_fail "[${name}] PUT /products/${product_id} request failed"
    return 1
  fi

  parse_http_response "$raw"

  if [[ "$HTTP_STATUS" != "200" ]]; then
    print_fail "[${name}] PUT /products/${product_id} expected 200, got ${HTTP_STATUS}"
    return 1
  fi

  if ! validate_name_and_price "$TEST_PRODUCT_NAME" "$UPDATED_PRODUCT_PRICE"; then
    print_fail "[${name}] PUT /products/${product_id} did not reflect updated values"
    return 1
  fi

  print_ok "[${name}] PUT /products/${product_id} passed"
  return 0
}

function test_delete_product() {
  local name=$1
  local base_url=$2
  local product_id=$3

  local raw
  if ! raw=$(http_request "DELETE" "${base_url}/products/${product_id}"); then
    print_fail "[${name}] DELETE /products/${product_id} request failed"
    return 1
  fi

  parse_http_response "$raw"

  if [[ "$HTTP_STATUS" != "204" ]]; then
    print_fail "[${name}] DELETE /products/${product_id} expected 204, got ${HTTP_STATUS}"
    return 1
  fi

  if ! raw=$(http_request "GET" "${base_url}/products/${product_id}"); then
    print_fail "[${name}] GET after delete for ${product_id} request failed"
    return 1
  fi
  parse_http_response "$raw"
  if [[ "$HTTP_STATUS" != "404" ]]; then
    print_fail "[${name}] GET deleted /products/${product_id} expected 404, got ${HTTP_STATUS}"
    return 1
  fi

  print_ok "[${name}] DELETE /products/${product_id} passed and resource gone"
  return 0
}

function record_summary() {
  local name=$1
  local status=$2
  local message=$3
  SERVICE_SUMMARY+=("${name}|${status}|${message}")
}

function print_summary() {
  print_info "===== Test Summary ====="
  local entry
  for entry in "${SERVICE_SUMMARY[@]}"; do
    IFS='|' read -r name status message <<<"$entry"
    if [[ "$status" == "PASS" ]]; then
      print_ok "${name}: ${message}"
    else
      print_fail "${name}: ${message}"
    fi
  done

  if (( FAILED_SERVICES == 0 )); then
    print_ok "All services passed"
  else
    print_fail "${FAILED_SERVICES} service(s) failed"
  fi
}

function run_service_tests() {
  local key=$1
  local name=$2
  local base_url=$3

  print_info "----- Testing ${name} (${base_url}) -----"

  if ! wait_for_service "$name" "$base_url"; then
    return 1
  fi

  if ! test_health "$name" "$base_url"; then
    return 1
  fi

  if ! test_create_product "$name" "$base_url"; then
    return 1
  fi

  local product_id="$LAST_PRODUCT_ID"

  if ! test_get_product "$name" "$base_url" "$product_id" "$TEST_PRODUCT_NAME" "$TEST_PRODUCT_PRICE"; then
    return 1
  fi

  if ! test_list_products "$name" "$base_url"; then
    return 1
  fi

  if ! test_update_product "$name" "$base_url" "$product_id"; then
    return 1
  fi

  if ! test_get_product "$name" "$base_url" "$product_id" "$TEST_PRODUCT_NAME" "$UPDATED_PRODUCT_PRICE" "after update"; then
    return 1
  fi

  if ! test_delete_product "$name" "$base_url" "$product_id"; then
    return 1
  fi

  return 0
}

SELECTED_SERVICES=()

function select_services() {
  local target=$1
  SELECTED_SERVICES=()
  local entry
  for entry in "${SERVICES[@]}"; do
    IFS='|' read -r key name <<<"$entry"
    if [[ "$target" == "all" || "$target" == "$key" ]]; then
      local base_url
      base_url=$(service_base_url "$key")
      if [[ -z "$base_url" ]]; then
        print_fail "Base URL for ${name} (${key}) is not configured"
        exit 1
      fi
      SELECTED_SERVICES+=("${key}|${name}|${base_url}")
    fi
  done
}

function main() {
  parse_arguments "$@"

  if [[ $JQ_AVAILABLE -eq 0 ]]; then
    print_info "jq not found - falling back to text-based JSON checks"
  fi

  configure_base_urls
  print_routing_info

  if ! ensure_proxy_health; then
    exit 1
  fi

  select_services "$TARGET"

  if [[ "${#SELECTED_SERVICES[@]}" -eq 0 ]]; then
    print_fail "No services matched target '${TARGET}'"
    exit 1
  fi

  local entry
  for entry in "${SELECTED_SERVICES[@]}"; do
    IFS='|' read -r key name url <<<"$entry"
    ((TOTAL_SERVICES++))
    if run_service_tests "$key" "$name" "$url"; then
      record_summary "$name" "PASS" "All tests passed"
    else
      ((FAILED_SERVICES++))
      record_summary "$name" "FAIL" "One or more tests failed"
    fi
  done

  print_summary

  if (( FAILED_SERVICES > 0 )); then
    exit 1
  fi
}

main "$@"
