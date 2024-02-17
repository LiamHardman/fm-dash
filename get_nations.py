nations = {}

with open("nationalities.txt", "r") as file:
    for line in file:
        parts = line.strip().split()
        if len(parts) >= 2:
            country = " ".join(parts[:-1])
            code = parts[-1]
            nations[country] = code
print(nations)
