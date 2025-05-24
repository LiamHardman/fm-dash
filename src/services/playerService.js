// src/services/playerService.js

// Ensure API_URL matches your Go backend configuration.
// If your Go backend is running on port 8091 and Vue dev server on 3000,
// Vite proxy in vite.config.js should handle requests to /upload, /api/players, /api/roles.
// For direct fetch calls from client to backend (if not using proxy for all), use full URL.
// Assuming proxy handles these paths.
const API_BASE_URL = ""; // Use relative paths if proxy is set up for all API routes

export default {
  /**
   * Upload a player file to the API for parsing.
   * @param {FormData} formData - The form data containing the file.
   * @returns {Promise<Object>} - A promise that resolves to an object
   * { datasetId: string, message: string, detectedCurrencySymbol: string }.
   */
  async uploadPlayerFile(formData) {
    try {
      // Path should match proxy if used, or be full URL if not.
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
      throw error; // Re-throw to be caught by the caller in the store
    }
  },

  /**
   * Fetches player data and the dataset's currency symbol from the backend.
   * @param {string} datasetId - The ID of the dataset to retrieve.
   * @param {string|null} position - Optional position filter (e.g., "DC").
   * @param {string|null} role - Optional role filter (e.g., "DC - Central Defender - Defend").
   * @returns {Promise<Object>} - A promise that resolves to an object
   * { players: Player[], currencySymbol: string }.
   */
  async getPlayersByDatasetId(datasetId, position = null, role = null) {
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
        // Roles can contain spaces and special characters, ensure they are encoded.
        params.append("role", role);
      }

      const queryString = params.toString();
      if (queryString) {
        url += `?${queryString}`;
      }

      console.log(`playerService: Fetching players from URL: ${url}`);

      const response = await fetch(url);

      if (!response.ok) {
        if (response.status === 404) {
          throw new Error(
            `Player data not found for ID: ${datasetId}. The data might have expired or the ID is incorrect.`,
          );
        }
        const errorText = await response.text();
        throw new Error(
          `API Error: ${response.status} - ${errorText || response.statusText}`,
        );
      }
      return await response.json(); // Expected: { players: [], currencySymbol: "€" }
    } catch (error) {
      console.error(
        "Error fetching players by dataset ID in playerService:",
        error,
      );
      throw error; // Re-throw
    }
  },

  /**
   * Fetches the list of all available role names from the backend.
   * @returns {Promise<string[]>} A promise that resolves to an array of role name strings.
   */
  async getAvailableRoles() {
    try {
      const response = await fetch(`${API_BASE_URL}/api/roles`);
      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(
          `API Error fetching roles: ${response.status} - ${errorText || response.statusText}`,
        );
      }
      return await response.json(); // Expected: ["Role Name 1", "Role Name 2", ...]
    } catch (error) {
      console.error("Error fetching available roles in playerService:", error);
      throw error; // Re-throw
    }
  },
};
