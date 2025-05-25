# Bedrock Model Validation Feature

## Overview

This feature enhances the bedrock-claude-proxy to not only read mappings from configuration files but also validate them against the actual available models in AWS Bedrock. This ensures that only valid, available models are exposed through the API.

## Key Features

### 1. Bedrock API Integration
- **Function**: `GetBedrockAvailableModels()`
- **Purpose**: Fetches the list of available foundation models from AWS Bedrock
- **Endpoint**: `https://bedrock.{region}.amazonaws.com/foundation-models`
- **Authentication**: Uses AWS v4 signature with existing credentials

### 2. Model Mapping Validation
- **Function**: `ValidateModelMappings()`
- **Purpose**: Cross-references configured model mappings with actual Bedrock availability
- **Returns**: Validation results showing which models are available/unavailable

### 3. Intelligent Model List Merging
- **Function**: `GetMergedModelList()`
- **Purpose**: Returns only models that are both configured AND available in Bedrock
- **Fallback**: If validation fails, falls back to configuration-only models

### 4. New API Endpoints

#### Enhanced `/v1/models` Endpoint
- Now returns only validated, available models
- Automatically filters out unavailable models
- Maintains backward compatibility

#### New `/v1/models/validate` Endpoint
- Provides detailed validation information
- Shows which models are available/unavailable
- Includes statistics and debugging information

## Implementation Details

### Data Structures

```go
// Represents a model from Bedrock API
type BedrockFoundationModel struct {
    ModelId                    string   `json:"modelId"`
    ModelName                  string   `json:"modelName"`
    ProviderName               string   `json:"providerName"`
    InputModalities            []string `json:"inputModalities"`
    OutputModalities           []string `json:"outputModalities"`
    ResponseStreamingSupported bool     `json:"responseStreamingSupported"`
}

// Validation result for each configured model
type ModelValidationResult struct {
    ConfigModel    string `json:"config_model"`
    BedrockModelId string `json:"bedrock_model_id"`
    IsValid        bool   `json:"is_valid"`
    Available      bool   `json:"available"`
}
```

### Error Handling

The implementation includes robust error handling:
- **API Failures**: Falls back to configuration-only models
- **Authentication Issues**: Logs warnings and continues with config models
- **Network Problems**: Graceful degradation with logging

### Logging

Enhanced logging provides visibility into:
- Model validation results
- API call success/failure
- Fallback behavior activation
- Unavailable model warnings

## Usage Examples

### 1. Get Available Models
```bash
curl -H "x-api-key: your-key" http://localhost:8080/v1/models
```

### 2. Validate Model Mappings
```bash
curl -H "x-api-key: your-key" http://localhost:8080/v1/models/validate
```

### 3. Example Validation Response
```json
{
  "validation_results": [
    {
      "config_model": "claude-3-opus-20240229",
      "bedrock_model_id": "anthropic.claude-3-opus-20240229-v1:0",
      "is_valid": true,
      "available": true
    },
    {
      "config_model": "claude-2.0",
      "bedrock_model_id": "anthropic.claude-v2",
      "is_valid": true,
      "available": false
    }
  ],
  "total_configured": 6,
  "available_count": 4,
  "unavailable_count": 2,
  "bedrock_models_count": 25
}
```

## Configuration

No additional configuration is required. The feature uses existing AWS credentials and settings:

```json
{
  "bedrock_config": {
    "access_key": "your-access-key",
    "secret_key": "your-secret-key",
    "region": "us-west-2",
    "model_mappings": {
      "claude-3-opus-20240229": "anthropic.claude-3-opus-20240229-v1:0",
      "claude-3-sonnet-20240229": "anthropic.claude-3-sonnet-20240229-v1:0"
    }
  }
}
```

## Benefits

1. **Reliability**: Only exposes models that are actually available
2. **Debugging**: Easy identification of configuration issues
3. **Monitoring**: Real-time validation of model availability
4. **Maintenance**: Automatic detection of deprecated models
5. **Compatibility**: Graceful fallback maintains service availability

## Technical Notes

### AWS Service Endpoints
- **Bedrock Service**: Used for listing foundation models
- **Bedrock Runtime**: Used for model invocation (unchanged)

### Performance Considerations
- Model validation is performed on-demand
- Results could be cached for better performance (future enhancement)
- Fallback ensures service remains responsive even during API issues

### Security
- Uses existing AWS IAM permissions
- Requires `bedrock:ListFoundationModels` permission for full functionality
- Gracefully handles permission issues

## Future Enhancements

1. **Caching**: Cache validation results to reduce API calls
2. **Periodic Validation**: Background validation with alerts
3. **Model Metadata**: Expose additional model information
4. **Health Checks**: Integration with monitoring systems
5. **Auto-Configuration**: Suggest optimal model mappings

## Troubleshooting

### Common Issues

1. **403 Forbidden**: Check AWS credentials and IAM permissions
2. **404 Not Found**: Verify region and endpoint configuration
3. **Fallback Behavior**: Check logs for validation failures

### Debug Mode

Enable debug mode in configuration to see detailed API interactions:
```json
{
  "bedrock_config": {
    "debug": true
  }
}
```

This will log all HTTP requests and responses for troubleshooting.
