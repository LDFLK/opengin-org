#!/bin/bash

# Load All Secretary Appointments Script
# Processes secretary appointment data for all presidents

set -e

# Color codes for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

print_header() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

# Load Gotabaya Rajapaksa secretaries
print_header "Loading Gotabaya Rajapaksa Secretary Appointments"

./orgchart -data "$(pwd)/data/people/Secretary/Gotabaya Rajapaksa/2019-12-16/2154-05/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Gotabaya Rajapaksa/2020-01-14/2158-19/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Gotabaya Rajapaksa/2020-01-24/2159-42/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Gotabaya Rajapaksa/2020-02-11/2162-62/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Gotabaya Rajapaksa/2020-02-14/2162-63/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Gotabaya Rajapaksa/2020-05-26/2177-07/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Gotabaya Rajapaksa/2020-08-13/2188-45/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Gotabaya Rajapaksa/2020-08-24/2190-13/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Gotabaya Rajapaksa/2020-12-08/2205-06/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Gotabaya Rajapaksa/2020-12-11/2205-15/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Gotabaya Rajapaksa/2020-12-16/2206-14/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Gotabaya Rajapaksa/2021-03-16/2219-33/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Gotabaya Rajapaksa/2021-05-04/2226-21/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Gotabaya Rajapaksa/2021-05-07/2226-61/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Gotabaya Rajapaksa/2021-05-17/2228-11/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Gotabaya Rajapaksa/2021-06-04/2190-13/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Gotabaya Rajapaksa/2021-06-11/2231-15/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Gotabaya Rajapaksa/2021-06-11/2231-22/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Gotabaya Rajapaksa/2021-07-07/2235-37/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Gotabaya Rajapaksa/2021-08-26/2242-4/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Gotabaya Rajapaksa/2021-10-08/2248-58/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Gotabaya Rajapaksa/2021-11-24/2255-19/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Gotabaya Rajapaksa/2021-12-26/2259-06/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Gotabaya Rajapaksa/2022-01-05/2261-34/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Gotabaya Rajapaksa/2022-01-10/2262-09/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Gotabaya Rajapaksa/2022-01-31/2265-19/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Gotabaya Rajapaksa/2022-04-26/2277-35/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Gotabaya Rajapaksa/2022-05-07/2278-28/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Gotabaya Rajapaksa/2022-05-11/2279-17/" -type secretary

print_success "Gotabaya Rajapaksa secretary appointments loaded"
echo ""

# Load Ranil Wickremesinghe secretaries
print_header "Loading Ranil Wickremesinghe Secretary Appointments"

./orgchart -data "$(pwd)/data/people/Secretary/Ranil Wickremasinghe/2022-05-17/2280-24/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Ranil Wickremasinghe/2022-05-30/2282-04/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Ranil Wickremasinghe/2022-06-17/2284-54/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Ranil Wickremasinghe/2022-07-07/2287-35/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Ranil Wickremasinghe/2022-11-14/2306-03/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Ranil Wickremasinghe/2022-11-25/2307-52/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Ranil Wickremasinghe/2023-03-30/2325-43/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Ranil Wickremasinghe/2023-04-20/2328-07/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Ranil Wickremasinghe/2023-11-02/2356-25/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Ranil Wickremasinghe/2023-11-23/2358-39/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Ranil Wickremasinghe/2023-11-23/2359-38/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Ranil Wickremasinghe/2023-12-08/2361-49/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Ranil Wickremasinghe/2024-01-11/2366-30/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Ranil Wickremasinghe/2024-04-10/2379-26/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Ranil Wickremasinghe/2024-04-10/2379-30/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Ranil Wickremasinghe/2024-05-03/2382-34/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Ranil Wickremasinghe/2024-06-07/2387-48/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Ranil Wickremasinghe/2024-06-07/2387-49/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Ranil Wickremasinghe/2024-06-27/2390-17/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/Ranil Wickremasinghe/2024-08-29/2399-42/" -type secretary

print_success "Ranil Wickremesinghe secretary appointments loaded"
echo ""

# Load Anura Kumara Dissanayake secretaries
print_header "Loading Anura Kumara Dissanayake Secretary Appointments"

./orgchart -data "$(pwd)/data/people/Secretary/AKD/2024-09-24/2403-12/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/AKD/2024-09-27/2403-51/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/AKD/2024-10-08/2405-14/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/AKD/2024-12-03/2413-25/" -type secretary
./orgchart -data "$(pwd)/data/people/Secretary/AKD/2024-12-13/2414-52/" -type secretary

print_success "Anura Kumara Dissanayake secretary appointments loaded"
echo ""

# Summary
print_header "Secretary Appointment Loading Complete"
echo -e "${GREEN}All secretary appointments have been processed!${NC}"
echo ""
echo "To verify the secretary appointments in the database, run:"
echo "docker exec neo4j cypher-shell -u neo4j -p neo4j123"
echo 'MATCH (m:Organisation)-[:SECRETARY_APPOINTED]->(s:Person) RETURN count(*) as SecretaryAppointments'
echo ""
