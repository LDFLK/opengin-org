#!/bin/bash
set -e

# Load Anura's secretary data
BASE="$(pwd)/data/secretaries/Anura Kumara Dissanayake"

./orgchart -data "$BASE/2024-09-27/2403-12"  -type secretary
./orgchart -data "$BASE/2024-09-27/2403-51"  -type secretary

./orgchart -data "$BASE/2024-10-03/2405-14"  -type secretary

./orgchart -data "$BASE/2024-11-19/2413-25-1"  -type secretary
./orgchart -data "$BASE/2024-11-20/2413-25-2"  -type secretary
./orgchart -data "$BASE/2024-11-21/2413-25-3"  -type secretary
./orgchart -data "$BASE/2024-11-25/2413-25-4"  -type secretary
./orgchart -data "$BASE/2024-11-28/2413-25-5"  -type secretary

./orgchart -data "$BASE/2024-12-03/2413-52-1"  -type secretary
./orgchart -data "$BASE/2024-12-11/2413-52-2"  -type secretary

./orgchart -data "$BASE/2025-03-06/2428-05"  -type secretary

./orgchart -data "$BASE/2025-05-21/2428-27"  -type secretary

./orgchart -data "$BASE/2025-06-23/2444-09"  -type secretary

./orgchart -data "$BASE/2025-10-10/2460-27-1"  -type secretary
./orgchart -data "$BASE/2025-10-14/2460-27-2"  -type secretary
./orgchart -data "$BASE/2025-10-16/2460-27-3"  -type secretary
