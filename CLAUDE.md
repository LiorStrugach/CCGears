# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Frigate is an NVR (Network Video Recorder) with real-time object detection for IP cameras, designed for Home Assistant integration. It uses OpenCV and TensorFlow for real-time object detection locally, with support for hardware accelerators like Google Coral.

## Tech Stack

- **Backend**: Python 3.x with FastAPI/Uvicorn
- **Frontend**: React 18 with TypeScript, Vite, TailwindCSS
- **Database**: SQLite with Peewee ORM, SqliteVec for embeddings
- **Video Processing**: FFmpeg, OpenCV
- **ML/AI**: TensorFlow, ONNX, support for various hardware accelerators (Coral, OpenVINO, TensorRT, etc.)
- **Communication**: MQTT, WebSocket, ZMQ
- **Storage**: SeaweedFS for S3-compatible storage
- **Container**: Docker with multi-architecture support

## Development Commands

### Backend Development

```bash
# Run tests
make run_tests

# Run Python unit tests only
docker run --rm --workdir=/opt/frigate --entrypoint= frigate:latest python3 -u -m unittest

# Run type checking with mypy
docker run --rm --workdir=/opt/frigate --entrypoint= frigate:latest python3 -u -m mypy --config-file frigate/mypy.ini frigate

# Run linting with ruff (configured in pyproject.toml)
ruff check frigate/
ruff format frigate/

# Build local Docker image
make local

# Run Frigate locally with Docker
make run
```

### Frontend Development

```bash
cd web/

npm install           # Install dependencies
npm run dev           # Development server (http://localhost:5173)
npm run build         # Production build
npm run lint          # Run linting
npm run lint:fix      # Auto-fix linting issues
npm run test          # Run tests (Vitest)
npm run coverage      # Tests with coverage
npm run prettier:write # Format code
```

### Docker Commands

```bash
make amd64            # Build for amd64
make arm64            # Build for arm64
make push             # Build and push to registry
docker-compose up     # Run with docker-compose
```

---

## System Architecture Overview

### Core Application Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                            FrigateApp (app.py)                               │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │   Camera    │  │   Object    │  │   Event     │  │    Embeddings       │ │
│  │   Capture   │──│  Detection  │──│  Processing │──│   (CLIP/Jina)       │ │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────────────┘ │
│         │                │                │                    │            │
│         └────────────────┴────────────────┴────────────────────┘            │
│                              │                                              │
│                    ┌─────────┴─────────┐                                    │
│                    │  Multiprocessing  │                                    │
│                    │     Queues        │                                    │
│                    └─────────┬─────────┘                                    │
│                              │                                              │
│  ┌──────────────────────────┴──────────────────────────┐                   │
│  │                    FastAPI (api/)                    │                   │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────────┐│                   │
│  │  │ Events  │ │ Camera  │ │ Agents  │ │   Review    ││                   │
│  │  │   API   │ │   API   │ │   API   │ │     API     ││                   │
│  │  └─────────┘ └─────────┘ └─────────┘ └─────────────┘│                   │
│  └──────────────────────────┬──────────────────────────┘                   │
└─────────────────────────────┼───────────────────────────────────────────────┘
                              │
              ┌───────────────┼───────────────┐
              │               │               │
        ┌─────┴─────┐  ┌──────┴─────┐  ┌──────┴──────┐
        │   MQTT    │  │  WebSocket │  │  WebPush    │
        │  Client   │  │   Client   │  │   Client    │
        └───────────┘  └────────────┘  └─────────────┘
```

### Directory Structure

```
frigate/
├── app.py                      # Main application orchestrator
├── models.py                   # Peewee ORM database models
├── const.py                    # Global constants
├── log.py                      # Logging infrastructure
├── api/                        # FastAPI routes
│   ├── fastapi_app.py         # App factory and middleware
│   ├── event.py               # Event endpoints
│   ├── camera.py              # Camera control
│   ├── agent.py               # AI agent endpoints (calls inserter.py)
│   ├── face.py                # Face recognition API
│   └── defs/                  # Request/response models
│       ├── query/             # Query parameter models
│       ├── request/           # Request body models
│       └── response/          # Response models
├── config/                     # Configuration system
│   ├── config.py              # Main FrigateConfig class
│   ├── base.py                # FrigateBaseModel
│   └── camera/                # Camera-specific configs
├── comms/                      # Inter-process communication
│   ├── dispatcher.py          # Message dispatcher
│   ├── base_communicator.py   # Abstract communicator
│   └── *_updater.py           # Various event updaters
├── agent/                      # Agent feature module
│   └── inserter.py            # Database CRUD operations
├── face/                       # Face recognition module
│   ├── face_collection/
│   │   └── inserter.py        # Collection CRUD
│   ├── face_assets/
│   │   └── inserter.py        # Asset CRUD
│   └── face_library/
│       └── inserter.py        # Library CRUD
├── vehicle/                    # Vehicle recognition module
│   ├── vehicle_collection/
│   │   └── inserter.py
│   └── vehicle_assets/
│       └── inserter.py
├── mentions/                   # Mention system
│   ├── mention_fetcher.py     # Fetch/resolve mentions
│   └── mention_updater.py     # Update mentions on changes
├── embeddings/                 # ML embeddings (CLIP, Jina)
├── events/                     # Event processing pipeline
├── detectors/                  # Hardware accelerator plugins
├── test/                       # Unit tests
│   └── http_api/              # API tests
├── util/                       # Shared utilities
└── migrations/                 # Database migrations

web/src/
├── api/                        # SWR + WebSocket setup
├── components/                 # React components
│   └── ui/                    # Radix UI wrappers
├── context/                    # React contexts
├── hooks/                      # Custom hooks
├── pages/                      # Page components
├── store/                      # Redux + RTK Query
├── types/                      # TypeScript types
└── utils/                      # Utility functions
```

---

## Coding Patterns and Best Practices

### CRITICAL: Inserter Pattern for CRUD Operations

**All database CRUD operations MUST be separated into `inserter.py` files.** API endpoints should NOT contain database logic directly.

**Feature Module Structure:**
```
frigate/
├── feature_name/
│   ├── __init__.py
│   └── inserter.py          # ALL database operations go here
└── api/
    └── feature_name.py      # API endpoints (calls inserter functions)
```

**Or for nested features:**
```
frigate/
├── face/
│   ├── face_collection/
│   │   └── inserter.py      # Collection CRUD
│   ├── face_assets/
│   │   └── inserter.py      # Asset CRUD
│   └── face_library/
│       └── inserter.py      # Library CRUD
└── api/
    └── face.py              # API endpoints
```

**Inserter Pattern Example (`agent/inserter.py`):**
```python
import logging
from uuid import uuid4
from peewee import prefetch
from frigate.models import Agent, db

logger = logging.getLogger(__name__)

# Private validation helpers
def _validate_agent_exists(agent_id: str) -> Agent:
    agent = Agent.get_or_none(Agent.id == agent_id)
    if not agent:
        logger.warning(f"Agent not found with id: {agent_id}")
        raise ValueError(f"Agent with id {agent_id} not found")
    return agent

def _validate_agent_name_exists(name: str, exclude_id: Optional[str] = None) -> bool:
    query = Agent.select().where(Agent.name == name)
    if exclude_id:
        query = query.where(Agent.id != exclude_id)
    return query.exists()

# CRUD Operations with @db.atomic() decorator
@db.atomic()
def insert_agent(
    positive_text_prompt: str,
    cameras: list[str],
    name: str,
    label: str,
    **kwargs
) -> Agent:
    try:
        agent = Agent.create(
            id=str(uuid4()),
            positive_text_prompt=positive_text_prompt,
            cameras=cameras,
            name=name,
            label=label,
            **kwargs
        )
        logger.info(f"Created agent with id: {agent.id}")
        return agent
    except Exception as e:
        logger.error(f"Failed to create agent: {e}", exc_info=True)
        raise

@db.atomic()
def get_agent_details_by_id(agent_id: str) -> dict:
    _validate_agent_exists(agent_id)
    query = prefetch(
        Agent.select().where(Agent.id == agent_id),
        NegativeTextPrompt.select(),
    )
    # ... build response dict
    return agent_data

@db.atomic()
def update_agent_data_by_id(agent_id: str, **update_fields) -> Agent:
    _validate_agent_exists(agent_id)
    update_fields = {k: v for k, v in update_fields.items() if v is not None}
    Agent.update(**update_fields).where(Agent.id == agent_id).execute()
    return Agent.get_by_id(agent_id)

@db.atomic()
def delete_agent_by_id(agent_id: str) -> bool:
    agent = _validate_agent_exists(agent_id)
    agent.delete_instance(recursive=True)
    logger.info(f"Deleted Agent with ID: {agent_id}")
    return True
```

**API Endpoint Pattern (`api/agent.py`):**
```python
from frigate.agent.inserter import (
    insert_agent,
    get_agent_details_by_id,
    update_agent_data_by_id,
    delete_agent_by_id,
    validate_camera_request,
)

router = APIRouter(tags=[Tags.agent])

@router.post("/agents", response_model=AgentDetailResponse, status_code=201)
def create_agent(request: Request, body: AgentCreateRequest):
    # Validation
    error_response = validate_camera_request(request, body.cameras)
    if error_response:
        return error_response

    # Call inserter function - NO direct database operations here
    agent = insert_agent(
        positive_text_prompt=body.positive_text_prompt,
        cameras=body.cameras,
        name=body.name,
        label=body.label,
    )

    return JSONResponse(
        content={"status": "success", "data": {"id": agent.id}},
        status_code=201,
    )

@router.get("/agents/{agent_id}", response_model=AgentDetailResponse)
def get_agent(agent_id: str):
    agent_data = get_agent_details_by_id(agent_id)
    if not agent_data:
        return JSONResponse(
            content={"status": "error", "message": "Agent not found"},
            status_code=404,
        )
    return JSONResponse(content={"status": "success", "data": agent_data})
```

### Pydantic Configuration Models

All configuration uses Pydantic v2 with strict validation:

```python
from pydantic import BaseModel, Field, field_validator, model_validator

class FrigateBaseModel(BaseModel):
    model_config = ConfigDict(extra="forbid", protected_namespaces=())

class CameraConfig(FrigateBaseModel):
    name: str = Field(title="Camera name")
    enabled: bool = Field(default=True)

    @field_validator("name")
    @classmethod
    def validate_name(cls, v):
        if not re.match(REGEX_CAMERA_NAME, v):
            raise ValueError("Invalid camera name")
        return v

    @model_validator(mode="after")
    def post_validation(self) -> Self:
        # Cross-field validation
        return self
```

### Request/Response Models

Define request bodies and responses as Pydantic models in `api/defs/`:

```python
# api/defs/request/events_body.py
class EventsSubLabelBody(BaseModel):
    subLabel: str = Field(title="Sub label", max_length=100)
    subLabelScore: Optional[float] = Field(default=None, gt=0.0, le=1.0)

# api/defs/query/events_query_parameters.py
class EventsQueryParams(BaseModel):
    camera: Optional[str] = "all"
    limit: Optional[int] = 100
    after: Optional[float] = None

# api/defs/response/event_response.py
class EventResponse(BaseModel):
    id: str
    label: str
    camera: str
    start_time: float
    data: dict[str, Any]
    model_config = ConfigDict(protected_namespaces=())
```

### Database Models (Peewee ORM)

```python
from peewee import Model, CharField, DateTimeField, ForeignKeyField, BooleanField
from playhouse.sqlite_ext import JSONField

class Event(Model):  # type: ignore[misc]
    id = CharField(null=False, primary_key=True, max_length=30)
    label = CharField(index=True, max_length=20)
    camera = CharField(index=True, max_length=20)
    start_time = DateTimeField()
    end_time = DateTimeField()
    data = JSONField()  # Semi-structured data

class AgentEvent(Model):  # type: ignore[misc]
    agent = ForeignKeyField(Agent, backref="agent_links")
    event = ForeignKeyField(Event, backref="event_links")

    class Meta:
        primary_key = CompositeKey("agent", "event")
```

### Decorator Patterns

**Database Transaction Decorator:**
```python
from frigate.models import db

@db.atomic()  # Wraps function in database transaction
def insert_agent(...):
    ...
```

**Post-Action Callback Decorator:**
```python
from frigate.util.decorators import run_after

@run_after(update_face_collection_mentions)  # Runs callback after success
def insert_face_collection(id: str, name: str):
    ...
```

### Abstract Base Classes for Extensibility

```python
from abc import ABC, abstractmethod

class Communicator(ABC):
    """Abstract pub/sub communicator."""

    @abstractmethod
    def publish(self, topic: str, payload: Any, retain: bool = False) -> None:
        pass

    @abstractmethod
    def subscribe(self, receiver: Callable) -> None:
        pass

# Implementations
class MqttClient(Communicator): ...
class WebSocketClient(Communicator): ...
```

### Logging Pattern

```python
import logging

logger = logging.getLogger(__name__)

# Usage throughout the file
logger.info(f"Created agent with id: {agent.id}")
logger.warning(f"Agent not found with id: {agent_id}")
logger.error(f"Failed to create agent: {e}", exc_info=True)
logger.debug(f"Processing frame: {frame_id}")
```

---

## Frontend Patterns

### Component Structure

```tsx
interface AutoUpdatingCameraImageProps {
  camera: string;
  reloadInterval?: number;
}

export default function AutoUpdatingCameraImage({
  camera,
  reloadInterval = 1000,
}: AutoUpdatingCameraImageProps) {
  const [key, setKey] = useState(Date.now());

  const handleLoad = useCallback(() => {
    // Logic here
  }, [dependencies]);

  return <img src={`/api/camera/${camera}?cache=${key}`} onLoad={handleLoad} />;
}
```

### State Management Layers

```tsx
// Layer 1: SWR for API data
const { data: config } = useSWR<FrigateConfig>("config");

// Layer 2: Redux + RTK Query for complex state
const { data: agentsData } = useGetAgentsQuery({ "X-CSRF-TOKEN": 1 });
const agents = useMemo(() => agentsData?.data?.agents || [], [agentsData]);

// Layer 3: WebSocket for real-time updates
export function useDetectState(camera: string) {
  const { value: { payload }, send } = useWs(
    `${camera}/detect/state`,
    `${camera}/detect/set`
  );
  return { payload, send };
}

// Layer 4: Context for UI state
const { theme, setTheme } = useTheme();
```

### Styling with Tailwind + CVA

```tsx
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";

const buttonVariants = cva(
  "inline-flex items-center justify-center rounded-md text-sm font-medium",
  {
    variants: {
      variant: {
        default: "bg-primary text-primary-foreground hover:bg-primary/80",
        secondary: "bg-background text-secondary hover:text-primary",
      },
    },
    defaultVariants: { variant: "default" },
  }
);

export function Button({ className, variant, ...props }: ButtonProps) {
  return (
    <button className={cn(buttonVariants({ variant }), className)} {...props} />
  );
}
```

---

## Python Unit Test Guidelines

### Test Organization

```
frigate/test/
├── __init__.py
├── const.py                    # Test constants (TEST_DB paths)
├── test_config.py              # Configuration tests
├── test_inserter_functions.py  # Inserter function tests
├── http_api/
│   ├── __init__.py
│   ├── base_http_test.py       # Base class for API tests
│   ├── test_http_event.py      # Event API tests
│   ├── test_http_agent.py      # Agent API tests
│   └── test_http_face.py       # Face API tests
```

### Base Test Class Pattern

```python
import unittest
from unittest.mock import MagicMock, patch
from fastapi.testclient import TestClient

class BaseTestHttp(unittest.TestCase):
    def setUp(self, models):
        """Initialize database for each test."""
        migrate_db = SqliteExtDatabase(TEST_DB)
        router = Router(migrate_db)
        router.run()

        self.db = SqliteQueueDatabase(TEST_DB)
        self.db.bind(models)

        self.minimal_config = {
            "mqtt": {"host": "mqtt"},
            "auth": {"enabled": False},
            "cameras": {
                "front_door": {
                    "ffmpeg": {"inputs": [{"path": "rtsp://...", "roles": ["detect"]}]},
                    "detect": {"height": 1080, "width": 1920, "fps": 5},
                }
            },
        }

    def tearDown(self):
        if not self.db.is_closed():
            self.db.close()
        for file in TEST_DB_CLEANUPS:
            try:
                os.remove(file)
            except OSError:
                pass

    def create_app(self, stats=None):
        return create_fastapi_app(
            FrigateConfig(**self.minimal_config),
            self.db, None, None, None, None, None, None, stats, None, None, None
        )
```

### Writing API Tests

```python
class TestAgentApi(BaseTestHttp):
    def setUp(self):
        super().setUp([Agent, NegativeTextPrompt, Event, Rules])
        self.app = super().create_app()

        # Override FastAPI dependencies
        dep = router.dependencies[0].dependency if router.dependencies else None
        if dep:
            self.app.dependency_overrides[dep] = lambda: None

        self.client = TestClient(self.app)
        self.test_agent_id = "test-agent-id-123"

    def create_mock_agent(self, agent_id=None):
        """Factory method for test data."""
        return Agent(
            id=agent_id or self.test_agent_id,
            status="Active",
            name="Test Agent",
            label="person",
            cameras=["front_door"],
        )

    def test_create_agent_success(self):
        with (
            patch("frigate.api.agent.insert_agent") as mock_insert,
            patch("frigate.api.agent.validate_camera_request") as mock_validate,
        ):
            mock_insert.return_value = self.create_mock_agent()
            mock_validate.return_value = None

            response = self.client.post("/agents", json={
                "positive_text_prompt": "positive prompt",
                "cameras": ["front_door"],
                "name": "Test Agent",
                "label": "person",
            })

            assert response.status_code == 201
            assert response.json()["status"] == "success"
            mock_insert.assert_called_once()

    def test_get_agent_not_found(self):
        with patch("frigate.api.agent.get_agent_details_by_id") as mock_get:
            mock_get.return_value = None

            response = self.client.get(f"/agents/{self.test_agent_id}")

            assert response.status_code == 404
            assert response.json()["status"] == "error"
```

### Test Naming Conventions

```python
# Pattern: test_<action>_<condition>_<expected_result>
def test_create_agent_success(self): ...
def test_create_agent_invalid_camera(self): ...
def test_get_agent_by_id_not_found(self): ...
def test_delete_agent_success(self): ...
```

---

## Adding New Features

### 1. Adding a New CRUD Feature (e.g., "widgets")

**Step 1: Create the feature module with inserter:**
```
frigate/
└── widget/
    ├── __init__.py
    └── inserter.py
```

**`widget/inserter.py`:**
```python
import logging
from uuid import uuid4
from frigate.models import Widget, db

logger = logging.getLogger(__name__)

def _validate_widget_exists(widget_id: str) -> Widget:
    widget = Widget.get_or_none(Widget.id == widget_id)
    if not widget:
        raise ValueError(f"Widget with id {widget_id} not found")
    return widget

@db.atomic()
def insert_widget(name: str, **kwargs) -> Widget:
    widget = Widget.create(id=str(uuid4()), name=name, **kwargs)
    logger.info(f"Created widget: {widget.id}")
    return widget

@db.atomic()
def get_widget_by_id(widget_id: str) -> Widget:
    return _validate_widget_exists(widget_id)

@db.atomic()
def update_widget(widget_id: str, **fields) -> Widget:
    _validate_widget_exists(widget_id)
    Widget.update(**fields).where(Widget.id == widget_id).execute()
    return Widget.get_by_id(widget_id)

@db.atomic()
def delete_widget(widget_id: str) -> bool:
    widget = _validate_widget_exists(widget_id)
    widget.delete_instance()
    return True
```

**Step 2: Create request/response models in `api/defs/`:**
```python
# api/defs/request/widget_body.py
class WidgetCreateBody(BaseModel):
    name: str = Field(max_length=100)

# api/defs/response/widget_response.py
class WidgetResponse(BaseModel):
    success: bool
    data: dict
```

**Step 3: Create the API router:**
```python
# api/widget.py
from frigate.widget.inserter import (
    insert_widget, get_widget_by_id, update_widget, delete_widget
)

router = APIRouter(tags=[Tags.widget])

@router.post("/widgets", response_model=WidgetResponse, status_code=201)
def create_widget(body: WidgetCreateBody):
    widget = insert_widget(name=body.name)
    return JSONResponse(content={"success": True, "data": {"id": widget.id}}, status_code=201)
```

**Step 4: Register in `api/fastapi_app.py`:**
```python
from frigate.api import widget
app.include_router(widget.router)
```

**Step 5: Add tests in `test/http_api/test_http_widget.py`**

### 2. Adding a New Database Model

**Step 1: Define the model in `models.py`:**
```python
class Widget(Model):  # type: ignore[misc]
    id = CharField(primary_key=True, max_length=36)
    name = CharField(max_length=255)
    created_at = DateTimeField(constraints=[SQL("DEFAULT CURRENT_TIMESTAMP")])
```

**Step 2: Create a migration in `migrations/`:**
```python
def migrate(migrator, database, fake=False, **kwargs):
    migrator.sql('''
        CREATE TABLE IF NOT EXISTS "widget" (
            "id" VARCHAR(36) PRIMARY KEY,
            "name" VARCHAR(255) NOT NULL,
            "created_at" DATETIME DEFAULT CURRENT_TIMESTAMP
        )
    ''')
```

---

## Important Paths

| Path | Description |
|------|-------------|
| `/config/config.yml` | Main configuration file |
| `/config/frigate.db` | SQLite database |
| `/media/frigate/recordings/` | Video recordings |
| `/media/frigate/clips/` | Event clips |
| `/config/model_cache/` | ML model cache |
| `/opt/frigate/web/` | Frontend build |

## Port Usage

| Port | Service |
|------|---------|
| 5000 | Internal NGINX |
| 5001 | Frigate API (FastAPI) |
| 5173 | Vite dev server |
| 8554 | go2rtc RTSP |
| 8555 | go2rtc WebRTC |
| 8888 | SeaweedFS S3 |
| 10002 | Frigate Notifications |
| 1883 | MQTT broker |

---

## Key Conventions Summary

1. **Inserter Pattern**: ALL database CRUD operations in `inserter.py` files, NOT in API endpoints
2. **Type hints**: Comprehensive type hints in Python; TypeScript required for frontend
3. **Logging**: Use `logger = logging.getLogger(__name__)` pattern
4. **Transactions**: Use `@db.atomic()` decorator for database operations
5. **Validation**: Private `_validate_*` helpers in inserter files
6. **Error handling**: Catch specific exceptions, return JSONResponse with status codes
7. **API responses**: Always include `success` boolean and descriptive `message`
8. **Testing**: Inherit from `BaseTestHttp`, mock inserter functions in API tests
9. **UUIDs**: Generate with `str(uuid4())` for all new entity IDs
10. **Frontend state**: Use SWR for API data, WebSocket for real-time, Redux for complex state
- Always look over the styling of the app and match the design to the rest of the components
- CRUCIAL: When creating a plan for a task, include break down of the task into sub-agents