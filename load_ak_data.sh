#!/bin/bash

# Load Anura's people and presidency gazettes
./orgchart -data "$(pwd)/data/documents/Anura Kumara Dissanayake/person/" -type document

# Load Anura's org gazettes
./orgchart -data "$(pwd)/data/documents/Anura Kumara Dissanayake/organisation/" -type document

# Load Anura's presidency data (termination of Ranil's presidency and starting Anura's)
./orgchart -data "$(pwd)/data/people/Anura Kumara Dissanayake/2024-09-23/2403-03-1" -type person # Add Anura as president
./orgchart -data "$(pwd)/data/orgchart/Anura Kumara Dissanayake/2024-09-23/" # move all Ranil's ministries to Anura
./orgchart -data "$(pwd)/data/people/Anura Kumara Dissanayake/2024-09-23/2403-03-2" -type person # terminate all Ranil's old people, assign everything to Anura

./orgchart -data "$(pwd)/data/people/Anura Kumara Dissanayake/2024-09-25/2403-37" -type person

# terminate all the old Ranil's portfolios which were transferred to Anura and all the people assigned (all Anura)
./orgchart -data "$(pwd)/data/people/Anura Kumara Dissanayake/2024-09-25/2403-38-1" -type person # terminate Anura assigned to all the old mins
./orgchart -data "$(pwd)/data/orgchart/Anura Kumara Dissanayake/2024-09-25/2403-38-1" # terminate all old depts and mins from Ranil

# Add Anura's new cabinet
./orgchart -data "$(pwd)/data/orgchart/Anura Kumara Dissanayake/2024-09-25/2403-38-2" # add some ministries
./orgchart -data "$(pwd)/data/orgchart/Anura Kumara Dissanayake/2024-09-25/2403-39" # add some more ministers

# Add Anura's new cabinet people
./orgchart -data "$(pwd)/data/people/Anura Kumara Dissanayake/2024-09-25/2403-38-2" -type person # assign people to the new ministries
./orgchart -data "$(pwd)/data/people/Anura Kumara Dissanayake/2024-09-25/2403-39" -type person # assign people to the new ministers


./orgchart -data "$(pwd)/data/orgchart/Anura Kumara Dissanayake/2024-09-27/"

# load the rest of Anura's org data
./orgchart -data "$(pwd)/data/orgchart/Anura Kumara Dissanayake/2024-11-18/2411-09"
./orgchart -data "$(pwd)/data/orgchart/Anura Kumara Dissanayake/2024-11-18/2411-10"


# Load Anura's people data
./orgchart -data "$(pwd)/data/people/Anura Kumara Dissanayake/2024-11-18/2411-09/" -type person
./orgchart -data "$(pwd)/data/people/Anura Kumara Dissanayake/2024-11-18/2411-10/" -type person

./orgchart -data "$(pwd)/data/orgchart/Anura Kumara Dissanayake/2024-11-25"

#!! AKD's latest data in 2025
./orgchart -data "$(pwd)/data/orgchart/Anura Kumara Dissanayake/2025-10-11/" -type organisation
./orgchart -data "$(pwd)/data/people/Anura Kumara Dissanayake/2025-10-11/" -type person

./orgchart -data "$(pwd)/data/orgchart/Anura Kumara Dissanayake/2025-10-18/" -type organisation

