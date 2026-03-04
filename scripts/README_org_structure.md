The following steps create the organisation structure for each minister. 
Specifically, they need to be done before inserting the secretary data if you are using any data backup before and including 0.0.5
1. Run `go run scripts/create_org_structure/main.go` - This will add the Organisation -> Minister -> Secretary structure for each minister.
2. Run `go run scripts/link_minister_roles/main.go` - This will link the existing people to the minister roles.

Alternatively you can run the script `./create_org_structure.sh` which will do the above steps.

To insert secretary data, you can then run:
`./load_ak_sec_data.sh`