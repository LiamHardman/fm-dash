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
              protoName: 'role_name',
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
              protoName: 'transfer_value',
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
              protoName: 'media_handling',
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
              protoName: 'nationality_fifa_code',
            },
            attributeMasked: {
              type: 'bool',
              id: 14,
              protoName: 'attribute_masked',
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
              protoName: 'parsed_positions',
            },
            shortPositions: {
              rule: 'repeated',
              type: 'string',
              id: 30,
              protoName: 'short_positions',
            },
            positionGroups: {
              rule: 'repeated',
              type: 'string',
              id: 31,
              protoName: 'position_groups',
            },
            essentialAttributes: {
              keyType: 'string',
              type: 'string',
              id: 32,
              protoName: 'essential_attributes',
            },
            bestRoleOverall: {
              type: 'string',
              id: 33,
              protoName: 'best_role_overall',
            },
            transferValueAmount: {
              type: 'int64',
              id: 34,
              protoName: 'transfer_value_amount',
            },
            wageAmount: {
              type: 'int64',
              id: 35,
              protoName: 'wage_amount',
            },
            totalStats: {
              type: 'int32',
              id: 36,
              protoName: 'total_stats',
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
              protoName: 'numeric_attributes',
            },
            performanceStatsNumeric: {
              keyType: 'string',
              type: 'double',
              id: 42,
              protoName: 'performance_stats_numeric',
            },
            performancePercentiles: {
              keyType: 'string',
              type: 'PerformancePercentileMap',
              id: 43,
              protoName: 'performance_percentiles',
            },
            basedIn: {
              type: 'string',
              id: 44,
              protoName: 'based_in',
            },
            saveCa: {
              type: 'int32',
              id: 45,
              protoName: 'save_ca',
            },
            pa: {
              type: 'int32',
              id: 46,
            },
            hiddenAttributes: {
              keyType: 'string',
              type: 'int32',
              id: 47,
              protoName: 'hidden_attributes',
            },
            inf: {
              type: 'string',
              id: 48,
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
              protoName: 'currency_symbol',
            },
            cacheData: {
              type: 'string',
              id: 3,
              protoName: 'cache_data',
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
              protoName: 'api_version',
            },
            fromCache: {
              type: 'bool',
              id: 3,
              protoName: 'from_cache',
            },
            requestId: {
              type: 'string',
              id: 4,
              protoName: 'request_id',
            },
            totalCount: {
              type: 'int32',
              id: 5,
              protoName: 'total_count',
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
              protoName: 'per_page',
            },
            totalPages: {
              type: 'int32',
              id: 3,
              protoName: 'total_pages',
            },
            totalCount: {
              type: 'int32',
              id: 4,
              protoName: 'total_count',
            },
            hasNext: {
              type: 'bool',
              id: 5,
              protoName: 'has_next',
            },
            hasPrevious: {
              type: 'bool',
              id: 6,
              protoName: 'has_previous',
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
              protoName: 'currency_symbol',
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
              protoName: 'error_code',
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
