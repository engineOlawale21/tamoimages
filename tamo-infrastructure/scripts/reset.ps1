[CmdletBinding(SupportsShouldProcess, ConfirmImpact = 'High')]
param([switch]$Force)

$ErrorActionPreference = 'Stop'
$composeFile = Join-Path (Split-Path -Parent $PSScriptRoot) 'docker-compose.yml'
$target = 'Tamo local PostgreSQL, Redis, and MinIO named-volume data'

Write-Warning 'This permanently deletes all local Tamo infrastructure data and recreates the containers.'
if (-not $Force -and -not $PSCmdlet.ShouldProcess($target, 'Delete and recreate')) {
    return
}

docker compose --file $composeFile down --volumes --remove-orphans
docker compose --file $composeFile up --detach
docker compose --file $composeFile ps --all
