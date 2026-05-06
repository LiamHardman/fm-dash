/**
 * Protobuf JSON descriptor for the fmdash API.
 *
 * AUTO-GENERATED — do not edit by hand.
 * Regenerate with: npm run generate:proto
 *
 * Sources:
 *   src/api/proto/player.proto
 *   src/api/proto/api_responses.proto
 */

const descriptor = {
  nested: {
    player: {
      options: {
        go_package: 'api/proto',
      },
      nested: {
        RoleOverallScore: {
          fields: {
            roleName: {
              type: 'string',
              id: 1,
            },
            score: {
              type: 'int32',
              id: 2,
            },
          },
        },
        Player: {
          fields: {
            uid: {
              type: 'int64',
              id: 1,
            },
            name: {
              type: 'string',
              id: 2,
            },
            position: {
              type: 'string',
              id: 3,
            },
            age: {
              type: 'string',
              id: 4,
            },
            club: {
              type: 'string',
              id: 5,
            },
            division: {
              type: 'string',
              id: 6,
            },
            transferValue: {
              type: 'string',
              id: 7,
            },
            wage: {
              type: 'string',
              id: 8,
            },
            personality: {
              type: 'string',
              id: 9,
            },
            mediaHandling: {
              type: 'string',
              id: 10,
            },
            nationality: {
              type: 'string',
              id: 11,
            },
            nationalityIso: {
              type: 'string',
              id: 12,
            },
            nationalityFifaCode: {
              type: 'string',
              id: 13,
            },
            attributeMasked: {
              type: 'bool',
              id: 14,
            },
            pac: {
              type: 'int32',
              id: 15,
            },
            sho: {
              type: 'int32',
              id: 16,
            },
            pas: {
              type: 'int32',
              id: 17,
            },
            dri: {
              type: 'int32',
              id: 18,
            },
            def: {
              type: 'int32',
              id: 19,
            },
            phy: {
              type: 'int32',
              id: 20,
            },
            gk: {
              type: 'int32',
              id: 21,
            },
            div: {
              type: 'int32',
              id: 22,
            },
            han: {
              type: 'int32',
              id: 23,
            },
            ref: {
              type: 'int32',
              id: 24,
            },
            kic: {
              type: 'int32',
              id: 25,
            },
            spd: {
              type: 'int32',
              id: 26,
            },
            pos: {
              type: 'int32',
              id: 27,
            },
            overall: {
              type: 'int32',
              id: 28,
            },
            parsedPositions: {
              rule: 'repeated',
              type: 'string',
              id: 29,
            },
            shortPositions: {
              rule: 'repeated',
              type: 'string',
              id: 30,
            },
            positionGroups: {
              rule: 'repeated',
              type: 'string',
              id: 31,
            },
            essentialAttributes: {
              keyType: 'string',
              type: 'string',
              id: 32,
            },
            bestRoleOverall: {
              type: 'string',
              id: 33,
            },
            transferValueAmount: {
              type: 'int64',
              id: 34,
            },
            wageAmount: {
              type: 'int64',
              id: 35,
            },
            totalStats: {
              type: 'int32',
              id: 36,
            },
            mbr: {
              type: 'int32',
              id: 37,
            },
            roleSpecificOveralls: {
              rule: 'repeated',
              type: 'RoleOverallScore',
              id: 38,
            },
            ca: {
              type: 'int32',
              id: 39,
            },
            attributes: {
              keyType: 'string',
              type: 'string',
              id: 40,
            },
            numericAttributes: {
              keyType: 'string',
              type: 'int32',
              id: 41,
            },
            performanceStatsNumeric: {
              keyType: 'string',
              type: 'double',
              id: 42,
            },
            performancePercentiles: {
              keyType: 'string',
              type: 'PerformancePercentileMap',
              id: 43,
            },
          },
        },
        PerformancePercentileMap: {
          fields: {
            percentiles: {
              keyType: 'string',
              type: 'double',
              id: 1,
            },
          },
        },
        DatasetData: {
          fields: {
            players: {
              rule: 'repeated',
              type: 'Player',
              id: 1,
            },
            currencySymbol: {
              type: 'string',
              id: 2,
            },
            cacheData: {
              type: 'string',
              id: 3,
            },
          },
        },
      },
    },
    api: {
      options: {
        go_package: 'api/proto',
      },
      nested: {
        ResponseMetadata: {
          fields: {
            timestamp: {
              type: 'int64',
              id: 1,
            },
            apiVersion: {
              type: 'string',
              id: 2,
            },
            fromCache: {
              type: 'bool',
              id: 3,
            },
            requestId: {
              type: 'string',
              id: 4,
            },
            totalCount: {
              type: 'int32',
              id: 5,
            },
          },
        },
        PaginationInfo: {
          fields: {
            page: {
              type: 'int32',
              id: 1,
            },
            perPage: {
              type: 'int32',
              id: 2,
            },
            totalPages: {
              type: 'int32',
              id: 3,
            },
            totalCount: {
              type: 'int32',
              id: 4,
            },
            hasNext: {
              type: 'bool',
              id: 5,
            },
            hasPrevious: {
              type: 'bool',
              id: 6,
            },
          },
        },
        PlayerDataResponse: {
          fields: {
            players: {
              rule: 'repeated',
              type: 'player.Player',
              id: 1,
            },
            currencySymbol: {
              type: 'string',
              id: 2,
            },
            metadata: {
              type: 'ResponseMetadata',
              id: 3,
            },
            pagination: {
              type: 'PaginationInfo',
              id: 4,
            },
          },
        },
        RolesResponse: {
          fields: {
            roles: {
              rule: 'repeated',
              type: 'string',
              id: 1,
            },
            metadata: {
              type: 'ResponseMetadata',
              id: 2,
            },
          },
        },
        LeaguesResponse: {
          fields: {
            leagues: {
              rule: 'repeated',
              type: 'string',
              id: 1,
            },
            metadata: {
              type: 'ResponseMetadata',
              id: 2,
            },
          },
        },
        TeamsResponse: {
          fields: {
            teams: {
              rule: 'repeated',
              type: 'string',
              id: 1,
            },
            metadata: {
              type: 'ResponseMetadata',
              id: 2,
            },
          },
        },
        SearchResponse: {
          fields: {
            players: {
              rule: 'repeated',
              type: 'player.Player',
              id: 1,
            },
            query: {
              type: 'string',
              id: 2,
            },
            metadata: {
              type: 'ResponseMetadata',
              id: 3,
            },
            pagination: {
              type: 'PaginationInfo',
              id: 4,
            },
          },
        },
        ErrorResponse: {
          fields: {
            errorCode: {
              type: 'string',
              id: 1,
            },
            message: {
              type: 'string',
              id: 2,
            },
            details: {
              rule: 'repeated',
              type: 'string',
              id: 3,
            },
            metadata: {
              type: 'ResponseMetadata',
              id: 4,
            },
          },
        },
        GenericResponse: {
          fields: {
            data: {
              type: 'string',
              id: 1,
            },
            metadata: {
              type: 'ResponseMetadata',
              id: 2,
            },
          },
        },
      },
    },
  },
}

export default descriptor
