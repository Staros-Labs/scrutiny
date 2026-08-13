# SMART Host Management

Scrutiny groups SMART devices by non-empty collector `host_id` values. Open **Hosts** from desktop navigation or **Settings > SMART Host Management** on mobile.

The host page does not include ZFS pools, MDADM arrays, Btrfs filesystems, filesystem-capacity records, or devices whose collector did not send a host ID.

## SMART dashboard navigation

The SMART dashboard pages by host instead of by device. All SMART devices assigned to one host remain together on the same page. Dashboard Settings provides 5, 10, 25, and 50 hosts-per-page options; the default is 10.

Use **Search hosts** on the SMART dashboard to filter host IDs across the full active or archived result set. Search is case-insensitive and runs before pagination.

Archiving, unarchiving, or deleting a device reloads the current page. If that page no longer exists, Scrutiny loads the last valid page. MDADM arrays remain on the dedicated MDADM page.

## Archive and unarchive

Archiving a host marks every SMART device currently assigned to that host as archived. Device metadata and time-series history remain stored. Unarchiving restores every device assigned to that host.

Use archive when retiring a collector temporarily or when historical data must remain available.

## Permanent purge

Purge deletes selected SMART device rows and their matching InfluxDB history. It cannot be undone.

Before purging:

1. Stop collectors on every selected host.
2. Confirm selected host IDs and device counts.
3. Type `PURGE` in the confirmation dialog.

An active collector can re-register devices immediately after purge. Purge does not alter data on physical drives.

### Shared WWN protection

InfluxDB history is keyed by WWN, while device metadata is keyed by `device_id`. If any WWN belonging to a selected host is also used by a device outside the full selected host set, Scrutiny blocks purge for the affected host rather than deleting shared history.

Select every intended host sharing that WWN, or resolve incorrect device identity before retrying.

### Partial failures and retries

The API returns a result for every requested host. Successfully purged hosts are removed from selection; failed hosts remain selected and can be retried.

For each host, Scrutiny deletes InfluxDB history before deleting SQLite device rows. If history deletion fails, the device rows remain, so repeating the purge is safe. If SQLite deletion fails after history deletion, repeating the operation safely retries idempotent history deletion before the SQLite delete.

Review the per-host error before retrying. Storage or connectivity failures should be corrected first; shared-WWN blocks require an identity or selection change.

## API

- `GET /api/hosts`
- `POST /api/hosts/archive`
- `POST /api/hosts/unarchive`
- `POST /api/hosts/purge`

See [openapi.yaml](openapi.yaml) for request and response schemas.
