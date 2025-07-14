# This was just a one-time script to run in order to get the JSON containing all the team ID's to team name mappings
# to be able to show logos if I wanted to at some point in this project.
# Will output a JSON file to the same directory as the script.
# Credit to 95brkz on sortitoutsi for this PDF!
import json
import re
import PyPDF2 # Added for PDF reading
import os # Added to construct file path

def pdf_to_json_teams(pdf_text_content):
    """
    Parses text content extracted from a PDF to create a JSON object
    mapping Team IDs to Team Names.

    Args:
        pdf_text_content (str): The full text content extracted from the PDF.

    Returns:
        str: A JSON string representing the Team ID to Team Name mapping.
    """
    teams_data = {}
    lines = pdf_text_content.splitlines()

    for line_number, raw_line in enumerate(lines):
        line = raw_line.strip()

        # Skip empty lines, page markers, and obvious header/summary lines
        if not line or line.startswith("--- PAGE") or " Clubs" in line.lower() or line.count('/') > 1 :
            parts_for_header_check = line.split()
            if len(parts_for_header_check) > 1 and parts_for_header_check[1].isalpha() and len(parts_for_header_check[1]) > 3:
                continue 
            if " Clubs" in line.lower():
                 continue

        # Attempt to parse format: "COUNTRY_CODE","ID","TEAM_NAME"
        if line.startswith('"') and '","' in line:
            raw_parts = line.split('","')
            if len(raw_parts) == 3:
                try:
                    country_code = raw_parts[0].strip().lstrip('"').strip()
                    team_id_str = raw_parts[1].strip()
                    team_name_str = raw_parts[2].strip().rstrip('"').strip()

                    team_id_cleaned = team_id_str.rstrip('-')
                    team_name_cleaned = team_name_str.rstrip('-').strip()

                    if (len(country_code) == 3 and country_code.isalpha() and country_code.isupper() and
                            team_id_cleaned.isdigit()):
                        if team_id_cleaned and team_name_cleaned: 
                            teams_data[team_id_cleaned] = team_name_cleaned
                        continue 
                except IndexError:
                    pass 

        # Attempt to parse format: COUNTRY_CODE ID TEAM_NAME
        parts = line.split()
        if len(parts) >= 3:
            country_code = parts[0]
            team_id_str = parts[1]
            
            team_id_cleaned = team_id_str.rstrip('-')

            if (len(country_code) == 3 and country_code.isalpha() and country_code.isupper() and
                    team_id_cleaned.isdigit()):
                team_name_parts = parts[2:]
                team_name = " ".join(team_name_parts).strip()
                team_name_cleaned = team_name.rstrip('-').strip()

                if team_id_cleaned and team_name_cleaned:
                    teams_data[team_id_cleaned] = team_name_cleaned
                continue

    return json.dumps(teams_data, indent=4)

if __name__ == '__main__':
    pdf_file_name = "fm24_teamid.pdf"
    # Get the directory where the script is located
    script_dir = os.path.dirname(os.path.abspath(__file__))
    pdf_file_path = os.path.join(script_dir, pdf_file_name)

    extracted_text = ""
    try:
        with open(pdf_file_path, 'rb') as pdf_file_obj:
            pdf_reader = PyPDF2.PdfReader(pdf_file_obj)
            num_pages = len(pdf_reader.pages)
            print(f"Reading '{pdf_file_name}'...")
            for page_num in range(num_pages):
                page_obj = pdf_reader.pages[page_num]
                extracted_text += page_obj.extract_text()
                print(f"Extracted text from page {page_num + 1}/{num_pages}")
        
        if not extracted_text.strip():
            print(f"Warning: No text could be extracted from '{pdf_file_name}'. The PDF might be image-based or scanned.")
            print("If the PDF is image-based, you'll need an OCR (Optical Character Recognition) tool to extract text first.")
        else:
            print("Text extraction complete. Parsing data...")
            json_output = pdf_to_json_teams(extracted_text)
            
            output_json_file = "teams_data.json"
            output_json_path = os.path.join(script_dir, output_json_file)

            with open(output_json_path, "w", encoding="utf-8") as f:
                f.write(json_output)
            print(f"\nSuccessfully parsed data and saved to '{output_json_file}' in the script's directory.")
            # Optionally print a snippet of the JSON
            # print("\n--- JSON Output Snippet ---")
            # print(json_output[:500] + "..." if len(json_output) > 500 else json_output)


    except FileNotFoundError:
        print(f"Error: The file '{pdf_file_name}' was not found in the script's directory: {script_dir}")
        print("Please ensure the PDF file is in the same directory as the script and named correctly.")
    except Exception as e:
        print(f"An error occurred: {e}")
        print("This could be due to an issue with the PDF file (e.g., encrypted, corrupted) or the PyPDF2 library.")

