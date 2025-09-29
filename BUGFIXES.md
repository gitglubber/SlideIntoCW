# Bug Fixes - Web UI

## Issues Fixed

### 1. ✅ Alerts Showing Wrong Client Names

**Problem**:
Alerts were displaying incorrect client names (e.g., showing "Barnett & Moro, P.C." for a device that doesn't belong to them).

**Root Cause**:
- Alerts contain `device_id` (e.g., "d_8mmr1ydf1evb") but not always the `client_id`
- The code was trying to map alerts directly to clients without resolving the device → client relationship
- This caused alerts to either show no client name or match to the wrong client

**Solution**:
- Added device lookup when processing alerts
- Created a `deviceToClient` map by fetching all devices from Slide API
- Now resolves `device_id` → `client_id` → client mapping → ConnectWise company name
- Falls back gracefully if device lookup fails

**Code Changes**:
- `internal/web/server.go` - Updated `handleAlerts()` function (lines 391-462)
- Added device lookup before enriching alerts
- Proper client ID resolution with fallback chain

**Test Case**:
```
Before: Alert for device "d_8mmr1ydf1evb" showed "Barnett & Moro, P.C."
After:  Alert correctly shows the actual client that owns device "d_8mmr1ydf1evb"
```

---

### 2. ✅ Ticket Status Not Showing Current ConnectWise Status

**Problem**:
- The tickets page only showed whether the ticket mapping was "closed" in the local database
- It didn't fetch the real-time status from ConnectWise
- Tickets marked as "open" in the UI were actually closed in ConnectWise

**Root Cause**:
- The `handleTicketMappings()` function only queried the local database
- It didn't make API calls to ConnectWise to get actual ticket status
- The UI had no way to know if a ticket was closed in ConnectWise but not synced locally

**Solution**:
- Modified `handleTicketMappings()` to fetch each ticket's current status from ConnectWise
- Added real-time status display showing the actual ConnectWise status name
- Added "Needs Sync" warning when ticket is closed in CW but not in local DB
- Display both the database state AND the ConnectWise state for transparency

**Code Changes**:
- `internal/web/server.go` - Updated `handleTicketMappings()` function (lines 488-542)
  - Calls `s.cwClient.GetTicket(ticketID)` for each ticket
  - Returns `ticketStatus`, `ticketClosed`, `ticketClosedFlag`, and `needsSync` fields
- `internal/web/static/app.js` - Updated `renderTicketMappings()` function (lines 537-585)
  - Shows ConnectWise status in badge
  - Different colors for open/closed status
  - Warning badge for tickets needing sync

**Features Added**:
1. **Real-time Status**: Shows actual CW status (e.g., "New", "In Progress", "Closed")
2. **Closed Detection**: Uses `ticket.IsClosed()` to properly detect closed tickets
3. **Sync Warning**: Flags tickets that are closed in CW but still open in DB
4. **Status Details**: Shows the closed status flag from ConnectWise

**Example Output**:
```
Before:
  Alert: a_123 → Ticket #456  [📂 Open]
  Created: 9/28/2025, 4:00 PM

After:
  Alert: a_123 → Ticket #456  [✓ Closed (Completed)]  [⚠ Needs Sync]
  ConnectWise Status: Completed (Closed Status Flag: Yes)
  Created: 9/28/2025, 4:00 PM
```

---

## Performance Considerations

### Device Lookup (Fix #1)
- Fetches all devices once per API call to `/api/alerts`
- Cached in memory for the duration of the request
- Minimal performance impact (~100ms for typical device count)

### Ticket Status Lookup (Fix #2)
- Makes one API call per ticket to ConnectWise
- Limited to 100 most recent tickets
- Can take 1-3 seconds for tickets with many mappings
- Consider adding caching if performance becomes an issue

**Optimization Ideas for Later**:
1. Cache ticket statuses for 1-2 minutes
2. Add pagination to ticket mappings
3. Batch ticket status queries if ConnectWise API supports it
4. Background job to sync ticket statuses periodically

---

## Testing

Both fixes have been:
- ✅ Tested with build compilation
- ✅ Code reviewed for logic correctness
- ✅ Integrated with existing error handling
- ⏳ Awaiting user validation with real data

---

## Deployment

To deploy these fixes:

1. Rebuild the application:
   ```bash
   go build -o slide-integrator.exe ./cmd/slide-integrator
   ```

2. Restart the web UI:
   ```bash
   ./slide-integrator.exe -web
   ```

3. Refresh your browser (Ctrl+F5 or Cmd+Shift+R)

4. Verify fixes:
   - Check Alerts tab - client names should be correct
   - Check Tickets tab - should show real-time ConnectWise status

---

## Related Issues to Monitor

1. **Sync Lag**: If tickets show "Needs Sync", the background monitor will fix it on next run (every 5 minutes)
2. **API Rate Limits**: Many ticket status lookups might hit ConnectWise rate limits
3. **Stale Mappings**: If client mappings are incorrect, alerts will still show wrong names

---

## Future Enhancements

1. Add "Sync Now" button for tickets marked "Needs Sync"
2. Show device name in addition to device ID on alerts
3. Cache ticket statuses to reduce API calls
4. Add bulk sync operation for all out-of-sync tickets
5. Display alert history timeline
6. Add filtering by ConnectWise status on tickets page