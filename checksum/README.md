# Checksum

This module caclulates the checksum of a remote file.

## CalculateChecksum

Calculate the checksum of a remote file using sha256sum. The module will retrieve
the file from the remote location and calculate the checksum.

### Parameters
- `url` (string) - The URL of the file to calculate the checksum for.

### Returns
- `string` - The checksum of the file.