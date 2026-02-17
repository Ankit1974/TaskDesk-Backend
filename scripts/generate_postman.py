import json
import re
import os

def generate_postman_collection():
    router_path = 'internal/api/router/router.go'
    if not os.path.exists(router_path):
        print(f"Error: {router_path} not found.")
        return

    with open(router_path, 'r') as f:
        content = f.read()

    # Find the group prefix
    group_match = re.search(r'r\.Group\("([^"]+)"\)', content)
    base_path = group_match.group(1) if group_match else ""

    # Find endpoints
    # Matches: api.GET("/health", handlers.HealthCheck)
    endpoints = re.findall(r'api\.([A-Z]+)\("([^"]+)",\s+handlers\.([A-Za-z]+)\)', content)

    collection = {
        "info": {
            "name": "TaskDesk Backend",
            "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
        },
        "item": []
    }

    for method, path, handler in endpoints:
        full_path = f"{base_path}{path}".strip("/")
        path_parts = full_path.split("/")
        
        item = {
            "name": f"{handler} ({method})",
            "request": {
                "method": method,
                "header": [],
                "url": {
                    "raw": "{{base_url}}/" + full_path,
                    "host": ["{{base_url}}"],
                    "path": path_parts
                }
            },
            "response": []
        }
        
        # Add a default body for POST requests
        if method == "POST":
            item["request"]["body"] = {
                "mode": "raw",
                "raw": "{\n  \"example\": \"data\"\n}",
                "options": {
                    "raw": {
                        "language": "json"
                    }
                }
            }

        collection["item"].append(item)

    output_path = 'TaskDesk_Backend.postman_collection.json'
    with open(output_path, 'w') as f:
        json.dump(collection, f, indent=4)
    
    print(f"Postman collection generated at: {output_path}")

if __name__ == "__main__":
    generate_postman_collection()
