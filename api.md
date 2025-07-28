# 📘 RepProject API – DNS Endpoints

Base URL: `https://repproject.world`

Free Token: `@repproject`

All endpoints require **Bearer token authentication** using the `Authorization` header unless stated otherwise.

```
Authorization: Bearer <token>
```

---

## 🔍 GET `/api/dns` ( free, limited results upto 1k )

Query DNS records (A, AAAA, NS, MX, TXT, CNAME) using a single parameter.

### Query Parameters

| Name    | Type   | Description                   | Required |
| ------- | ------ | ----------------------------- | -------- |
| `ip`    | string | IPv4 or domain for A records  | optional |
| `ipv6`  | string | IPv6 address for AAAA records | optional |
| `ns`    | string | Domain to query NS records    | optional |
| `mx`    | string | Domain to query MX records    | optional |
| `txt`   | string | Domain to query TXT records   | optional |
| `cname` | string | Domain to query CNAME records | optional |

> Only one query parameter should be supplied.

### Example Request

```
GET /api/dns?ip=1.1.1.1
```

### Success Response

```json
[
  {
    "ip": "192.168.1.1",
    "domain_id": "abc123.co,",
    "record_type": "A",
    "timestamp": 1724457600
  },
  {
    "ip": "2606:4700:4700::1111",
    "domain_id": "def456.com",
    "record_type": "AAAA",
    "timestamp": 1724458600
  }
]
```

### Errors

- `400 Bad Request` – If no valid query param is provided.
- `404 Not Found` – DNS record not found.

---

## 📄 GET `/api/dns/paging`

Paginated DNS record querying.

### Query Parameters

Same as `/api/dns`, plus:

| Name         | Type   | Description                             | Required |
| ------------ | ------ | --------------------------------------- | -------- |
| `page_size`  | int    | Number of results per page (default 10) | optional |
| `page_token` | string | Base64-encoded paging state             | optional |

### Example Request

```
GET /api/dns/paging?ip=8.8.8.8&page_size=50
```

### Success Response

```json
{
  "data": [
    {
      "ip": "8.8.8.8",
      "domain_id": "example123.com",
      "record_type": "A",
      "timestamp": 1724457600
    },
    {
      "ip": "8.8.8.8",
      "domain_id": "another.com",
      "record_type": "A",
      "timestamp": 1724457700
    }
  ],
  "pagination": {
    "page_size": 50,
    "has_more": true,
    "next_page_token": "eyJvZmZzZXQiOjUwfQ..."
  }
}
```

---

## 🔁 GET `/api/dns/a`

**Reverse A Record Lookup** (for IPv4). This is a **paid feature**.

### Query Parameters

| Name         | Type   | Description                     | Required |
| ------------ | ------ | ------------------------------- | -------- |
| `ipv4`       | string | IPv4 address to look up         | yes      |
| `page_size`  | int    | Results per page (default 10)   | optional |
| `page_token` | string | Base64-encoded pagination token | optional |

### Example

```
GET /api/dns/a?ipv4=1.1.1.1&page_size=25
```

### Success Response

```json
{
  "data": [
    {
      "domain_id": "cloudflare-dns.com",
      "ip": "1.1.1.1",
      "asn": 13335,
      "asn_name": "CLOUDFLARENET",
      "country": "US",
      "city": "San Francisco",
      "latlong": "37.7749,-122.4194",
      "timestamp": 1724457600
    }
    ...
  ],
  "pagination": {
    "page_size": 25,
    "has_more": false
  }
}

```

### Errors

- `401 Unauthorized` – Free users cannot access this endpoint.
- `404 Not Found`

---

## 🔁 GET `/api/dns/aaaa`

**Reverse AAAA Record Lookup** (for IPv6). This is a **paid feature**.

### Query Parameters

| Name         | Type   | Description                     | Required |
| ------------ | ------ | ------------------------------- | -------- |
| `ipv6`       | string | IPv6 address to look up         | yes      |
| `page_size`  | int    | Results per page (default 10)   | optional |
| `page_token` | string | Base64-encoded pagination token | optional |

### Example

```
GET /api/dns/aaaa?ipv6=2606:4700:4700::1111&page_size=25
```

### Success Response

```json
{
  "data": [
    {
      "domain_id": "cloudflare-dns.com",
      "ip": "2606:4700:4700::1111",
      "asn": 13335,
      "asn_name": "CLOUDFLARENET",
      "country": "US",
      "city": "San Francisco",
      "latlong": "37.7749,-122.4194",
      "timestamp": 1724457600
    }
    ...
  ],
  "pagination": {
    "page_size": 25,
    "has_more": false
  }
}

```

### Errors

- `401 Unauthorized` – Free users cannot access this endpoint.
- `404 Not Found`

---
