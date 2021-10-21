import json
def get_arn():
    with open("iamResponse.json") as read_file:
        data = json.load(read_file)
        if data["Role"]:
            if data["Role"]["Arn"]:
                return data["Role"]["Arn"]

arn = get_arn()
if arn:
    print(arn)
