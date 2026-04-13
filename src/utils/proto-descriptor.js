/**
 * Protobuf JSON descriptor for the fmdash API.
 *
 * This file is the canonical JavaScript representation of the schemas defined in:
 *   src/api/proto/player.proto
 *   src/api/proto/api_responses.proto
 *
 * HOW TO REGENERATE
 * -----------------
 * Run the following after any change to the .proto files:
 *
 *   npm run generate:proto
 *
 * which executes:
 *   npx pbjs -t json src/api/proto/player.proto src/api/proto/api_responses.proto \
 *            -o src/utils/proto-descriptor.json
 *
 * and then re-exports it from this file.
 *
 * IMPORTANT: Keep this file in sync with the .proto sources. Field IDs and
 * types must match exactly or the binary decoder will silently drop or
 * misinterpret fields.
 */

const descriptor = {
  nested: {
    api: {
      nested: {
        ResponseMetadata: {
          fields: {
            timestamp: { type: 'int64', id: 1 },
            api_version: { type: 'string', id: 2 },
            from_cache: { type: 'bool', id: 3 },
            request_id: { type: 'string', id: 4 },
            total_count: { type: 'int32', id: 5 },
          },
        },
        PlayerDataResponse: {
          fields: {
            players: { rule: 'repeated', type: 'player.Player', id: 1 },
            currency_symbol: { type: 'string', id: 2 },
            metadata: { type: 'ResponseMetadata', id: 3 },
            pagination: { type: 'PaginationInfo', id: 4 },
          },
        },
        RolesResponse: {
          fields: {
            roles: { rule: 'repeated', type: 'string', id: 1 },
            metadata: { type: 'ResponseMetadata', id: 2 },
          },
        },
        GenericResponse: {
          fields: {
            data: { type: 'string', id: 1 },
            metadata: { type: 'ResponseMetadata', id: 2 },
          },
        },
        PaginationInfo: {
          fields: {
            page: { type: 'int32', id: 1 },
            per_page: { type: 'int32', id: 2 },
            total_pages: { type: 'int32', id: 3 },
            total_count: { type: 'int32', id: 4 },
            has_next: { type: 'bool', id: 5 },
            has_previous: { type: 'bool', id: 6 },
          },
        },
      },
    },
    player: {
      nested: {
        RoleOverallScore: {
          fields: {
            role_name: { type: 'string', id: 1 },
            score: { type: 'int32', id: 2 },
          },
        },
        Player: {
          fields: {
            uid: { type: 'int64', id: 1 },
            name: { type: 'string', id: 2 },
            position: { type: 'string', id: 3 },
            age: { type: 'string', id: 4 },
            club: { type: 'string', id: 5 },
            division: { type: 'string', id: 6 },
            transfer_value: { type: 'string', id: 7 },
            wage: { type: 'string', id: 8 },
            personality: { type: 'string', id: 9 },
            media_handling: { type: 'string', id: 10 },
            nationality: { type: 'string', id: 11 },
            nationalityIso: { type: 'string', id: 12 },
            nationality_fifa_code: { type: 'string', id: 13 },
            attribute_masked: { type: 'bool', id: 14 },

            // Essential FIFA-style stats for display and sorting
            pac: { type: 'int32', id: 15 },
            sho: { type: 'int32', id: 16 },
            pas: { type: 'int32', id: 17 },
            dri: { type: 'int32', id: 18 },
            def: { type: 'int32', id: 19 },
            phy: { type: 'int32', id: 20 },
            gk: { type: 'int32', id: 21 },
            div: { type: 'int32', id: 22 },
            han: { type: 'int32', id: 23 },
            ref: { type: 'int32', id: 24 },
            kic: { type: 'int32', id: 25 },
            spd: { type: 'int32', id: 26 },
            pos: { type: 'int32', id: 27 },
            overall: { type: 'int32', id: 28 },

            // Position data for filtering and grouping
            parsed_positions: { rule: 'repeated', type: 'string', id: 29 },
            short_positions: { rule: 'repeated', type: 'string', id: 30 },
            position_groups: { rule: 'repeated', type: 'string', id: 31 },

            // Essential attributes for display
            essential_attributes: { rule: 'map', keyType: 'string', type: 'string', id: 32 },

            // Best role for display
            best_role_overall: { type: 'string', id: 33 },

            // Numeric values for sorting
            transfer_value_amount: { type: 'int64', id: 34 },
            wage_amount: { type: 'int64', id: 35 },

            // Moneyball Rating and Total Stats
            total_stats: { type: 'int32', id: 36 },
            mbr: { type: 'int32', id: 37 },

            // Role-specific overall scores for tactical analysis
            roleSpecificOveralls: {
              rule: 'repeated',
              type: 'player.RoleOverallScore',
              id: 38,
            },

            // Cached Current Ability (id 39 — added in player.proto; must stay in sync)
            ca: { type: 'int32', id: 39 },
          },
        },
      },
    },
  },
}

export default descriptor
