// src/services/playerService.js
const API_BASE_URL = ""; // Use relative paths if proxy is set up for all API routes

export default {
  async uploadPlayerFile(formData) {
    try {
      const response = await fetch(`${API_BASE_URL}/upload`, {
        method: "POST",
        body: formData,
      });
      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(
          `API Error: ${response.status} - ${errorText || response.statusText}`,
        );
      }
      return await response.json();
    } catch (error) {
      console.error("Upload error in playerService:", error);
      throw error;
    }
  },

  async getPlayersByDatasetId(
    datasetId,
    position = null,
    role = null,
    ageRange = null,
    transferValueRange = null,
  ) {
    if (!datasetId) {
      return Promise.reject(new Error("Dataset ID is required."));
    }
    try {
      let url = `${API_BASE_URL}/api/players/${datasetId}`;
      const params = new URLSearchParams();
      if (position) {
        params.append("position", position);
      }
      if (role) {
        params.append("role", role);
      }
      if (ageRange) {
        if (ageRange.min !== null && ageRange.min !== undefined) {
          params.append("minAge", ageRange.min.toString());
        }
        if (ageRange.max !== null && ageRange.max !== undefined) {
          // Assuming your backend doesn't need maxAge if it's the slider's absolute max
          // If it does, send it. For now, let's assume only send if not default max.
          // This depends on your backend logic for "Any" max age.
          // For simplicity, let's send it if it's not the default "Any" value (e.g. 50)
          // This needs to align with how PlayerFilters defines "Any" for max age.
          // Let's assume if ageRange.max is less than a known upper bound (e.g. 50), it's a specific filter.
          // From PlayerFilters, ageSliderMax is 50. If filters.ageRange.max === ageSliderMax, it implies "Any".
          // So, only send maxAge if it's *not* the default max from the slider.
          // However, the backend will handle -1 as "not set".
          // The PlayerFilters component sends ageRange.min and ageRange.max directly.
          // The backend should interpret a missing param or -1 as "no filter for this bound".
          params.append("maxAge", ageRange.max.toString());
        }
      }
      if (transferValueRange) {
        if (
          transferValueRange.min !== null &&
          transferValueRange.min !== undefined
        ) {
          params.append("minTransferValue", transferValueRange.min.toString());
        }
        if (
          transferValueRange.max !== null &&
          transferValueRange.max !== undefined
        ) {
          // Similar logic for maxTransferValue, send if it's not the default "Any"
          // The backend should interpret a missing param or -1 as "no filter for this bound".
          params.append("maxTransferValue", transferValueRange.max.toString());
        }
      }

      const queryString = params.toString();
      if (queryString) {
        url += `?${queryString}`;
      }

      console.log(`playerService: Fetching players from URL: ${url}`);

      const response = await fetch(url);
      if (!response.ok) {
        if (response.status === 404) {
          throw new Error(`Player data not found for ID: ${datasetId}.`);
        }
        const errorText = await response.text();
        throw new Error(
          `API Error: ${response.status} - ${errorText || response.statusText}`,
        );
      }
      return await response.json();
    } catch (error) {
      console.error(
        "Error fetching players by dataset ID in playerService:",
        error,
      );
      throw error;
    }
  },

  async getAvailableRoles() {
    try {
      const response = await fetch(`${API_BASE_URL}/api/roles`);
      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(
          `API Error fetching roles: ${response.status} - ${errorText || response.statusText}`,
        );
      }
      return await response.json();
    } catch (error) {
      console.error("Error fetching available roles in playerService:", error);
      throw error;
    }
  },
};
