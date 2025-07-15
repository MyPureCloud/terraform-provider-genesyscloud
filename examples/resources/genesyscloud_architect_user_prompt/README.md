# Architect User Prompt Examples

This directory contains examples for the `genesyscloud_architect_user_prompt` resource, demonstrating various usage patterns including local filesystem and S3 file support.

## Examples

### Basic TTS Prompt
- **File**: `resource.tf` (basic_prompt)
- **Description**: Simple text-to-speech prompt without audio files
- **Use Case**: Quick prompts that don't require custom audio

### Local Audio Files
- **File**: `resource.tf` (local_audio_prompt)
- **Description**: Prompt using local audio files
- **Use Case**: When audio files are stored locally in your project

### S3 Audio Files
- **File**: `resource.tf` (s3_audio_prompt)
- **Description**: Prompt using audio files stored in Amazon S3
- **Use Case**: Centralized audio file management in S3
- **Requirements**: AWS credentials configured

### Mixed Local and S3 Files
- **File**: `resource.tf` (mixed_prompt)
- **Description**: Prompt using both local and S3 audio files
- **Use Case**: Gradual migration or hybrid file storage strategy

## File Support

The `genesyscloud_architect_user_prompt` resource supports:

### Local Files
- Standard paths: `/path/to/audio.wav`
- Relative paths: `./audio/welcome.wav`

### S3 Files
- S3 URI format: `s3://bucket-name/path/to/audio.wav`
- Alternative format: `s3a://bucket-name/path/to/audio.wav`

### AWS Credentials
S3 files use the standard AWS credential chain:
- Environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`)
- AWS credentials file (`~/.aws/credentials`)
- IAM roles (EC2 instance profiles, EKS service accounts)
- AWS SSO profiles

## Usage Notes

1. **File Content Hash**: Always use `filesha256()` function for the `file_content_hash` field to detect changes
2. **Language Codes**: Use standard language codes (e.g., `en-us`, `es-es`, `fr-fr`)
3. **Audio Formats**: Supported formats include WAV, MP3, and other common audio formats
4. **S3 Permissions**: Ensure your AWS credentials have read access to the S3 buckets and objects 