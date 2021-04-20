#!/usr/bin/env bash
# Usage:   ./e2e.sh ${RedisVersion} ${FixtureVersion} ${PasswordProtected} ${PasswordFromFile}
# Example: ./e2e.sh 6.0.6 5.0 1 0

set -x
set -o pipefail
set -o nounset

redis_version="$1"
fixture_version="$2"
require_pass="$3"
pass_from_file="$4"
offline_scrape="$5"

skip_re="^(redis_sentinel_build_info|redis_sentinel_exporter_build_info|redis_sentinel_info|redis_sentinel_used_cpu|redis_sentinel_exporter_last_scrape_duration_seconds|redis_sentinel_uptime_in_seconds|redis_sentinel_connections_received_total|redis_sentinel_net|redis_sentinel_instantaneous|redis_sentinel_process_id)"

echo "==> Redis $redis_version"

rm -rf "redis-${redis_version}" "redis-${redis_version}.tar.gz"

wget "http://download.redis.io/releases/redis-${redis_version}.tar.gz"
tar -zxvf "redis-${redis_version}.tar.gz"
cd "redis-${redis_version}"
make

nohup ./src/redis-server &

success="0"
for i in {1..5} ; do
  if ./src/redis-cli -p 6379 PING; then
    success="1"
    break
  fi
  sleep 1
done

if [[ $success == "0" ]]; then
  echo "Redis PING failed"
  exit 1
fi

if [[ $require_pass == "0" ]]; then
  cp ../test_data/sentinel.conf sentinel.conf
else
  cp ../test_data/sentinel-protected.conf sentinel.conf
fi

cat sentinel.conf
nohup ./src/redis-sentinel sentinel.conf &

success="0"
for i in {1..5} ; do
  if ./src/redis-cli -p 26379 PING; then
    success="1"
    break
  fi
  sleep 1
done

if [[ $success == "0" ]]; then
  echo "Redis Sentinel PING failed"
  exit 1
fi

cd ../

go build

if [[ $require_pass == "0" ]]; then
  nohup ./redis_sentinel_exporter --debug &
else
  if [[ $pass_from_file == "0" ]]; then
    nohup ./redis_sentinel_exporter --debug --sentinel.password=ABCD &
  else
    echo "ABCD" > /tmp/sentinel_password.txt
    nohup ./redis_sentinel_exporter --debug --sentinel.password-file=/tmp/sentinel_password.txt &
  fi
fi

wget --retry-connrefused --tries=5 -O - "127.0.0.1:9355/metrics"| grep "redis_" | grep -E -v "${skip_re}" > "e2e-output.txt"

diff -u \
  "test_data/e2e-output-v${fixture_version}.txt" \
  "e2e-output.txt"

if [[ $offline_scrape == "1" ]]; then
  kill -9 "$(cat /var/run/redis-sentinel.pid)"

  failed="0"
  for i in {1..5} ; do
    if ! ../src/redis-cli -p 26379 PING; then
      failed="1"
      break
    fi
    sleep 1
  done

  if [[ $failed == "0" ]]; then
    echo "Redis Sentinel up and running"
    exit 1
  fi

  wget --retry-connrefused --tries=5 -O - "127.0.0.1:9355/metrics"| grep "redis_" | grep -E -v "${skip_re}" > "e2e-output-offline.txt"
  diff -u \
    "test_data/e2e-output-v${fixture_version}-offline.txt" \
    "e2e-output-offline.txt"
fi
