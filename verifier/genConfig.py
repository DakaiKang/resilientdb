import json

verifier_config = {"f": 1,
"timer": 180
}

with open('verifier_config.json', 'w') as json_file:
  json.dump(verifier_config, json_file)

