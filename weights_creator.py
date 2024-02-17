import tkinter as tk
from tkinter import messagebox, simpledialog
from tkinter import font as tkFont
import yaml
import json
import os


def load_config():
    # Try to get the path from the environment variable
    config_path = os.getenv("FMD_CONF_LOCATION", "config.yml")

    # Read the configuration file from the path
    with open(config_path) as f:
        return yaml.safe_load(f)


yaml_config = load_config()


attributes = yaml_config["attributes"]


def load_weights_data():
    try:
        with open("weights_data.json", "r") as file:
            return json.load(file)
    except FileNotFoundError:
        return {}  # Return an empty dictionary if the file does not exist


weights_data = load_weights_data()

# The main weights dictionary
weights = {}


def submit():
    new_weights = {}
    for category in attributes.values():
        for full, abbr in category.items():
            value = entries[abbr].get()
            new_weights[abbr] = int(value) * 5 if value.strip() else 1

    dict_name = dict_name_entry.get()
    if not dict_name:
        messagebox.showerror("Error", "Please enter a dictionary name")
        return

    # Update the weights dictionary in weights_data module
    weights_data.weights[dict_name] = new_weights

    # Write the updated weights dictionary back to weights_data.py
    with open("weights_data.json", "w") as file:
        json.dump(weights_data.weights, file, indent=4)

    messagebox.showinfo("Success", f"Dictionary '{dict_name}' added successfully!")

    # Clear the form
    for full in attribute_names.keys():
        entries[full].delete(0, tk.END)
    dict_name_entry.delete(0, tk.END)


def clear_entries():
    for category in attributes.values():
        for abbr in category.values():
            entries[abbr].delete(0, tk.END)


def on_dict_select(event):
    widget = event.widget
    if widget.curselection():
        clear_entries()
        index = int(widget.curselection()[0])
        dict_name = widget.get(index)
        selected_weights = weights_data[dict_name]
        for category in attributes.values():
            for full, abbr in category.items():
                entry_value = selected_weights.get(abbr, "")
                if entry_value:
                    entry_value = int(entry_value) // 5
                entries[abbr].insert(0, entry_value)


position_order = [
    "GK",
    "DL",
    "DC",
    "DR",
    "WBL",
    "WBR",
    "DM",
    "ML",
    "MC",
    "MR",
    "AML",
    "AMC",
    "AMR",
    "ST",
]


def custom_sort_key(key):
    # Extract the position part of the key and handle compound positions
    position = key.split(" - ")[0].split("/")[0]
    position_index = (
        position_order.index(position)
        if position in position_order
        else len(position_order)
    )
    return position_index, key


def refresh_listbox():
    dict_listbox.delete(0, tk.END)  # Clear existing entries
    sorted_keys = sorted(weights_data.keys(), key=custom_sort_key)  # Sort the keys
    for dict_name in sorted_keys:
        dict_listbox.insert(tk.END, dict_name)


def save_weights():
    selected_dict_name = (
        dict_listbox.get(tk.ANCHOR) if dict_listbox.curselection() else None
    )

    dict_name = simpledialog.askstring(
        "Dictionary Name", "Enter the dictionary name:", initialvalue=selected_dict_name
    )
    if not dict_name:
        messagebox.showerror("Error", "No dictionary name was provided.")
        return

    # Gather new weights data from the entries
    new_weights = {}
    for category in attributes.values():
        for full, abbr in category.items():
            value = entries[abbr].get()
            new_weights[abbr] = int(value) * 5 if value.strip() else 1

    # Add or update the new weights in the weights_data
    weights_data[dict_name] = new_weights

    # Sort the weights_data
    sorted_keys = sorted(weights_data.keys(), key=custom_sort_key)
    # Update the weights_data with the sorted data
    sorted_weights_data = {key: weights_data[key] for key in sorted_keys}
    weights_data.clear()
    weights_data.update(sorted_weights_data)

    # Write the sorted weights data to the file
    with open("weights_data.json", "w") as file:
        json.dump(weights_data, file, indent=4)

    # Refresh the listbox to reflect the new order
    refresh_listbox()

    messagebox.showinfo("Success", f"Dictionary '{dict_name}' saved successfully!")
    clear_entries()


def delete_selected_weight():
    selected_dict_name = dict_listbox.get(tk.ANCHOR)
    if selected_dict_name:
        if messagebox.askyesno(
            "Delete", f"Are you sure you want to delete '{selected_dict_name}'?"
        ):
            del weights_data[selected_dict_name]

            with open("weights_data.json", "w") as file:
                json.dump(weights_data, file, indent=4)

            dict_listbox.delete(tk.ANCHOR)

            clear_entries()
            messagebox.showinfo("Deleted", f"'{selected_dict_name}' has been deleted.")


root = tk.Tk()
root.title("Weights Dictionary GUI")

listbox_frame = tk.Frame(root)
listbox_frame.pack(side=tk.LEFT, fill=tk.BOTH, padx=10, pady=10)

dict_listbox = tk.Listbox(listbox_frame)
dict_listbox.pack(side=tk.TOP, fill=tk.BOTH, expand=True)

for dict_name in weights_data.keys():
    dict_listbox.insert(tk.END, dict_name)
refresh_listbox()
dict_listbox.bind("<<ListboxSelect>>", on_dict_select)

new_weight_button = tk.Button(listbox_frame, text="New Weight", command=clear_entries)
new_weight_button.pack(side=tk.TOP, fill=tk.X)

entries_frame = tk.Frame(root)
entries_frame.pack(side=tk.LEFT, fill=tk.BOTH, expand=True)

entries = {}
column_counter = 0

standard_font = tkFont.Font(size=12)


header_font = tkFont.Font(weight="bold", size=14)

for category, attrs in attributes.items():
    tk.Label(
        entries_frame,
        text=f"{category.replace('_', ' ').title()}",
        font=header_font,
    ).grid(row=0, column=column_counter, columnspan=2, sticky="ew")
    row_counter = 1
    for full, abbr in attrs.items():
        tk.Label(entries_frame, text=full, font=standard_font).grid(
            row=row_counter, column=column_counter, sticky="e"
        )
        entry = tk.Entry(entries_frame)
        entry.grid(row=row_counter, column=column_counter + 1, sticky="ew")
        entries[abbr] = entry
        row_counter += 1

    column_counter += 2
max_rows = max(len(attrs) for attrs in attributes.values()) + 2

save_button = tk.Button(entries_frame, text="Save", command=save_weights)
save_button.grid(row=max_rows, column=0, columnspan=1, pady=10, sticky="ew")

row_counter += 1

delete_weight_button = tk.Button(
    entries_frame, text="Delete Weight", command=delete_selected_weight
)
delete_weight_button.grid(row=max_rows, column=1, columnspan=1, pady=10, sticky="ew")

root.mainloop()
