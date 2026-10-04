package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"testing"
	"time"
)

var accessToken = env("ACROCUIT_ACCESS_TOKEN", "")

type apiResult struct {
	status int
	header http.Header
	data   json.RawMessage
	raw    string
}

type idRef struct {
	Id int `json:"id"`
}

type breakerNode struct {
	Id              int    `json:"id"`
	UpstreamBreaker *idRef `json:"upstream_breaker"`
}

type deviceNode struct {
	Id      int    `json:"id"`
	Breaker *idRef `json:"breaker"`
}

type downstream struct {
	Breakers []breakerNode `json:"breakers"`
	Devices  []deviceNode  `json:"devices"`
}

func request(t *testing.T, method, path string, body any) apiResult {
	t.Helper()

	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		reader = bytes.NewReader(payload)
	}

	return send(t, method, path, "application/json", reader, nil)
}

func send(t *testing.T, method, path, contentType string, body io.Reader, headers map[string]string) apiResult {
	t.Helper()

	req, err := http.NewRequest(method, host+apiPath+path, body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", contentType)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	if identityToken != "" {
		req.Header.Set("X-Forwarded-Identity", identityToken)
	}
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}

	res, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		t.Skipf("%s %s: %v", method, path, err)
	}
	defer res.Body.Close()

	raw, _ := io.ReadAll(res.Body)
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	_ = json.Unmarshal(raw, &envelope)

	return apiResult{status: res.StatusCode, header: res.Header, data: envelope.Data, raw: string(raw)}
}

func expectStatus(t *testing.T, r apiResult, want int, action string) {
	t.Helper()
	if r.status != want {
		t.Fatalf("%s: status = %d, want %d (%s)", action, r.status, want, r.raw)
	}
}

func decode[T any](t *testing.T, r apiResult) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(r.data, &v); err != nil {
		t.Fatalf("decode data: %v (%s)", err, r.raw)
	}
	return v
}

func create(t *testing.T, path string, body map[string]any) int {
	t.Helper()
	r := request(t, http.MethodPost, path, body)
	expectStatus(t, r, http.StatusOK, "POST "+path)
	return decode[idRef](t, r).Id
}

func newSpaceGroup(t *testing.T) int {
	t.Helper()
	id := create(t, "/space-groups", map[string]any{"name": t.Name()})
	t.Cleanup(func() {
		request(t, http.MethodDelete, fmt.Sprintf("/space-groups/%d", id), nil)
	})
	return id
}

func newSpace(t *testing.T, spaceGroupId int) int {
	t.Helper()
	return create(t, fmt.Sprintf("/space-groups/%d/spaces", spaceGroupId), map[string]any{"name": "space"})
}

func newBreakerGroup(t *testing.T, spaceId int) int {
	t.Helper()
	return create(t, fmt.Sprintf("/spaces/%d/breaker-groups", spaceId), map[string]any{"name": "breaker group"})
}

func newBreaker(t *testing.T, breakerGroupId, upstreamBreakerId int) int {
	t.Helper()
	body := map[string]any{"name": "breaker"}
	if upstreamBreakerId != 0 {
		body["upstream_breaker_id"] = upstreamBreakerId
	}
	return create(t, fmt.Sprintf("/breaker-groups/%d/breakers", breakerGroupId), body)
}

func newDevice(t *testing.T, spaceId int) int {
	t.Helper()
	return create(t, fmt.Sprintf("/spaces/%d/devices", spaceId), map[string]any{"name": "device"})
}

func attach(t *testing.T, deviceId, breakerId int) apiResult {
	t.Helper()
	return request(t, http.MethodPost, fmt.Sprintf("/devices/%d/breakers", deviceId), map[string]any{"breaker_id": breakerId})
}

func mustAttach(t *testing.T, deviceId, breakerId int) {
	t.Helper()
	expectStatus(t, attach(t, deviceId, breakerId), http.StatusOK, fmt.Sprintf("attach device %d to breaker %d", deviceId, breakerId))
}

func getDownstream(t *testing.T, breakerId int) downstream {
	t.Helper()
	r := request(t, http.MethodGet, fmt.Sprintf("/breakers/%d/downstream", breakerId), nil)
	expectStatus(t, r, http.StatusOK, "GET downstream")
	return decode[downstream](t, r)
}

func getUpstream(t *testing.T, deviceId int) []breakerNode {
	t.Helper()
	r := request(t, http.MethodGet, fmt.Sprintf("/devices/%d/upstream", deviceId), nil)
	expectStatus(t, r, http.StatusOK, "GET upstream")
	return decode[[]breakerNode](t, r)
}

func breakerIds(breakers []breakerNode) map[int]bool {
	ids := make(map[int]bool, len(breakers))
	for _, b := range breakers {
		ids[b.Id] = true
	}
	return ids
}

func hasDeviceOnBreaker(devices []deviceNode, deviceId, breakerId int) bool {
	for _, d := range devices {
		if d.Id == deviceId && d.Breaker != nil && d.Breaker.Id == breakerId {
			return true
		}
	}
	return false
}

func expectGone(t *testing.T, paths ...string) {
	t.Helper()
	for _, path := range paths {
		expectStatus(t, request(t, http.MethodGet, path, nil), http.StatusNotFound, "GET "+path)
	}
}

func TestE2EBreakerUpstreamRejectsCycle(t *testing.T) {
	group := newBreakerGroup(t, newSpace(t, newSpaceGroup(t)))
	root := newBreaker(t, group, 0)
	middle := newBreaker(t, group, root)
	leaf := newBreaker(t, group, middle)

	r := request(t, http.MethodPatch, fmt.Sprintf("/breakers/%d/upstream", root), map[string]any{"upstream_breaker_id": leaf})
	expectStatus(t, r, http.StatusBadRequest, "set root upstream to its own descendant")
}

func TestE2EBreakerUpstreamRejectsSelfReference(t *testing.T) {
	group := newBreakerGroup(t, newSpace(t, newSpaceGroup(t)))
	breaker := newBreaker(t, group, 0)

	r := request(t, http.MethodPatch, fmt.Sprintf("/breakers/%d/upstream", breaker), map[string]any{"upstream_breaker_id": breaker})
	expectStatus(t, r, http.StatusBadRequest, "set breaker upstream to itself")
}

func TestE2EAttachDeviceRejectsDuplicate(t *testing.T) {
	space := newSpace(t, newSpaceGroup(t))
	breaker := newBreaker(t, newBreakerGroup(t, space), 0)
	device := newDevice(t, space)

	mustAttach(t, device, breaker)
	expectStatus(t, attach(t, device, breaker), http.StatusConflict, "attach the same breaker twice")
}

func TestE2EBreakerDownstreamIncludesDescendants(t *testing.T) {
	space := newSpace(t, newSpaceGroup(t))
	group := newBreakerGroup(t, space)
	root := newBreaker(t, group, 0)
	left := newBreaker(t, group, root)
	right := newBreaker(t, group, root)
	leftLeaf := newBreaker(t, group, left)
	device := newDevice(t, space)
	mustAttach(t, device, leftLeaf)
	mustAttach(t, device, right)

	tree := getDownstream(t, root)

	ids := breakerIds(tree.Breakers)
	if len(tree.Breakers) != 4 || !ids[root] || !ids[left] || !ids[right] || !ids[leftLeaf] {
		t.Fatalf("downstream breakers = %v, want root, left, right, leftLeaf", ids)
	}
	if !hasDeviceOnBreaker(tree.Devices, device, leftLeaf) || !hasDeviceOnBreaker(tree.Devices, device, right) {
		t.Fatalf("device attached to two breakers should appear under both: %+v", tree.Devices)
	}
}

func TestE2EDeviceUpstreamDeduplicatesSharedPath(t *testing.T) {
	space := newSpace(t, newSpaceGroup(t))
	group := newBreakerGroup(t, space)
	root := newBreaker(t, group, 0)
	shared := newBreaker(t, group, root)
	switchA := newBreaker(t, group, shared)
	switchB := newBreaker(t, group, shared)
	device := newDevice(t, space)
	mustAttach(t, device, switchA)
	mustAttach(t, device, switchB)

	upstream := getUpstream(t, device)

	ids := breakerIds(upstream)
	if len(upstream) != 4 || !ids[root] || !ids[shared] || !ids[switchA] || !ids[switchB] {
		t.Fatalf("upstream = %v (%d entries), want root, shared, switchA, switchB once each", ids, len(upstream))
	}
}

func TestE2EDeleteSpaceRemovesCrossSpaceReferences(t *testing.T) {
	spaceGroup := newSpaceGroup(t)
	kept := newSpace(t, spaceGroup)
	removed := newSpace(t, spaceGroup)
	root := newBreaker(t, newBreakerGroup(t, kept), 0)
	keptSwitch := newBreaker(t, newBreakerGroup(t, kept), root)
	removedGroup := newBreakerGroup(t, removed)
	removedSwitch := newBreaker(t, removedGroup, root)
	sharedDevice := newDevice(t, kept)
	removedDevice := newDevice(t, removed)
	mustAttach(t, sharedDevice, keptSwitch)
	mustAttach(t, sharedDevice, removedSwitch)
	mustAttach(t, removedDevice, removedSwitch)

	expectStatus(t, request(t, http.MethodDelete, fmt.Sprintf("/spaces/%d", removed), nil), http.StatusOK, "delete space")

	expectGone(t,
		fmt.Sprintf("/spaces/%d", removed),
		fmt.Sprintf("/breaker-groups/%d", removedGroup),
		fmt.Sprintf("/breakers/%d", removedSwitch),
		fmt.Sprintf("/devices/%d", removedDevice),
	)

	ids := breakerIds(getUpstream(t, sharedDevice))
	if len(ids) != 2 || !ids[root] || !ids[keptSwitch] {
		t.Fatalf("upstream after delete = %v, want root and keptSwitch", ids)
	}

	tree := getDownstream(t, root)
	if len(tree.Breakers) != 2 || !hasDeviceOnBreaker(tree.Devices, sharedDevice, keptSwitch) {
		t.Fatalf("downstream after delete = %+v, want root and keptSwitch with sharedDevice", tree)
	}
}

func TestE2EDeleteSpaceGroupCascades(t *testing.T) {
	spaceGroup := create(t, "/space-groups", map[string]any{"name": t.Name()})
	space := newSpace(t, spaceGroup)
	group := newBreakerGroup(t, space)
	breaker := newBreaker(t, group, 0)
	device := newDevice(t, space)
	mustAttach(t, device, breaker)

	expectStatus(t, request(t, http.MethodDelete, fmt.Sprintf("/space-groups/%d", spaceGroup), nil), http.StatusOK, "delete space group")

	expectGone(t,
		fmt.Sprintf("/space-groups/%d", spaceGroup),
		fmt.Sprintf("/spaces/%d", space),
		fmt.Sprintf("/breaker-groups/%d", group),
		fmt.Sprintf("/breakers/%d", breaker),
		fmt.Sprintf("/devices/%d", device),
	)
}

func pngImage(t *testing.T, size int) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, size, size))); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return buf.Bytes()
}

func uploadBackgroundImage(t *testing.T, spaceId int, field string, data []byte) apiResult {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile(field, "background")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	part.Write(data)
	writer.Close()
	return send(t, http.MethodPut, fmt.Sprintf("/spaces/%d/background-image", spaceId), writer.FormDataContentType(), &body, nil)
}

func backgroundImagePath(spaceId int) string {
	return fmt.Sprintf("/spaces/%d/background-image", spaceId)
}

func TestE2ESpaceBackgroundImageRoundTrip(t *testing.T) {
	space := newSpace(t, newSpaceGroup(t))
	expectStatus(t, request(t, http.MethodGet, backgroundImagePath(space), nil), http.StatusNotFound, "GET before upload")
	createdAt := decode[struct {
		UpdatedAt int64 `json:"background_image_updated_at"`
	}](t, request(t, http.MethodGet, fmt.Sprintf("/spaces/%d", space), nil)).UpdatedAt

	time.Sleep(5 * time.Millisecond)
	first := pngImage(t, 2)
	uploaded := uploadBackgroundImage(t, space, "image", first)
	expectStatus(t, uploaded, http.StatusOK, "upload")
	uploadedAt := decode[struct {
		UpdatedAt int64 `json:"background_image_updated_at"`
	}](t, uploaded).UpdatedAt
	if uploadedAt <= createdAt {
		t.Fatalf("background_image_updated_at = %d after upload, want > %d", uploadedAt, createdAt)
	}

	got := request(t, http.MethodGet, backgroundImagePath(space), nil)
	expectStatus(t, got, http.StatusOK, "GET after upload")
	if got.header.Get("Content-Type") != "image/png" || got.raw != string(first) {
		t.Fatalf("GET returned %q (%d bytes), want image/png (%d bytes)", got.header.Get("Content-Type"), len(got.raw), len(first))
	}
	etag := got.header.Get("ETag")
	if etag != fmt.Sprintf(`"%d"`, uploadedAt) {
		t.Fatalf("ETag = %q, want %q", etag, fmt.Sprintf(`"%d"`, uploadedAt))
	}
	notModified := send(t, http.MethodGet, backgroundImagePath(space), "", nil, map[string]string{"If-None-Match": etag})
	expectStatus(t, notModified, http.StatusNotModified, "GET with matching If-None-Match")

	second := pngImage(t, 4)
	expectStatus(t, uploadBackgroundImage(t, space, "image", second), http.StatusOK, "replace")
	if replaced := request(t, http.MethodGet, backgroundImagePath(space), nil); replaced.raw != string(second) {
		t.Fatalf("GET after replace returned %d bytes, want %d", len(replaced.raw), len(second))
	}

	expectStatus(t, request(t, http.MethodDelete, backgroundImagePath(space), nil), http.StatusOK, "delete")
	expectStatus(t, request(t, http.MethodGet, backgroundImagePath(space), nil), http.StatusNotFound, "GET after delete")
	expectStatus(t, request(t, http.MethodDelete, backgroundImagePath(space), nil), http.StatusNotFound, "delete again")
}

func TestE2ESpaceBackgroundImageRejectsInvalidUpload(t *testing.T) {
	space := newSpace(t, newSpaceGroup(t))

	expectStatus(t, uploadBackgroundImage(t, space, "image", []byte("plain text, not an image")), http.StatusUnsupportedMediaType, "upload non-image")
	expectStatus(t, uploadBackgroundImage(t, space, "file", pngImage(t, 2)), http.StatusBadRequest, "upload without image field")
	expectStatus(t, uploadBackgroundImage(t, space, "image", make([]byte, 11<<20)), http.StatusRequestEntityTooLarge, "upload over 10 MB")
	expectStatus(t, uploadBackgroundImage(t, 0, "image", pngImage(t, 2)), http.StatusNotFound, "upload to missing space")
	expectStatus(t, request(t, http.MethodGet, backgroundImagePath(space), nil), http.StatusNotFound, "GET after rejected uploads")
}
