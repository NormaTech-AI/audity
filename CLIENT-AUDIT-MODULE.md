# Client Audit Module Implementation

## Overview
Implemented a comprehensive client audit module where client users can view assigned frameworks and submit responses to compliance questions. The module includes role-based access control where POC users see all questions while stakeholders only see questions assigned to them.

## Features

### Role-Based Question Visibility
- **POC Users (poc_client, poc_internal)**: Can see ALL questions in the audit
- **Stakeholders**: Can only see questions specifically assigned to them
- Questions can be assigned to specific users via `question_assignments` table

### Submission Workflow
1. **Draft Mode**: Users can save answers as drafts (status: `in_progress`)
2. **Submit for Review**: Once satisfied, users submit answers for auditor review (status: `submitted`)
3. **Review States**: Auditors can approve, reject, or refer submissions
4. **Resubmission**: Rejected submissions can be updated and resubmitted

## Backend Implementation

### Database Schema

#### Tables Used

**Client-Specific Database Tables:**
- `audits` - Framework assignments to clients
- `questions` - Questions from compliance frameworks
- `question_assignments` - Question delegation to stakeholders
- `submissions` - Client answers and explanations
- `evidence` - Supporting files for submissions
- `comments` - Discussion threads on submissions

#### New SQL Queries

**File**: `/services/tenant-service/db/client-queries/questions.sql`

1. **ListQuestionsForUser** - Role-based question filtering
   ```sql
   -- POC users see all questions, stakeholders see only assigned questions
   SELECT DISTINCT q.*, s.*, qa.assigned_to
   FROM questions q
   LEFT JOIN submissions s ON s.question_id = q.id
   LEFT JOIN question_assignments qa ON qa.question_id = q.id
   WHERE q.audit_id = $1
     AND ($2 = true OR qa.assigned_to = $3)
   ```

2. **AssignQuestionToUser** - Assign questions to stakeholders
3. **UnassignQuestionFromUser** - Remove question assignments
4. **ListQuestionAssignments** - List all assignments for a question
5. **ListUserAssignments** - List all questions assigned to a user

### API Endpoints

**Base Path**: `/api/client-audit`

#### 1. List Audits
```http
GET /api/client-audit
Authorization: Cookie (auth_token)
Permission: audit:list
```

**Response:**
```json
[
  {
    "id": "uuid",
    "framework_id": "uuid",
    "framework_name": "SOC 2 Type II",
    "due_date": "2024-12-31",
    "status": "in_progress",
    "total_questions": 150,
    "answered_count": 75,
    "progress_percent": 50.0,
    "created_at": "2024-01-01T00:00:00Z"
  }
]
```

#### 2. Get Audit Detail with Questions
```http
GET /api/client-audit/:auditId
Authorization: Cookie (auth_token)
Permission: audit:read
```

**Response:**
```json
{
  "id": "uuid",
  "framework_id": "uuid",
  "framework_name": "SOC 2 Type II",
  "due_date": "2024-12-31",
  "status": "in_progress",
  "created_at": "2024-01-01T00:00:00Z",
  "questions": [
    {
      "id": "uuid",
      "section": "Access Control",
      "question_number": "1.1",
      "question_text": "Do you have a documented access control policy?",
      "question_type": "yes_no",
      "help_text": "Provide details about your access control policy",
      "is_mandatory": true,
      "display_order": 1,
      "submission_id": "uuid",
      "answer_value": "yes",
      "answer_text": null,
      "explanation": "We have a comprehensive access control policy...",
      "submission_status": "in_progress",
      "submitted_at": null,
      "submitted_by": null,
      "is_assigned_to_me": true
    }
  ]
}
```

#### 3. Save Submission (Draft)
```http
POST /api/client-audit/submissions
Authorization: Cookie (auth_token)
Permission: audit:submit
Content-Type: application/json

{
  "question_id": "uuid",
  "answer_value": "yes",
  "answer_text": "Optional text answer",
  "explanation": "Detailed explanation (required)"
}
```

**Response:**
```json
{
  "id": "uuid",
  "question_id": "uuid",
  "status": "in_progress",
  "message": "Submission saved successfully"
}
```

#### 4. Submit Answer for Review
```http
POST /api/client-audit/submissions/:submissionId/submit
Authorization: Cookie (auth_token)
Permission: audit:submit
```

**Response:**
```json
{
  "id": "uuid",
  "status": "submitted",
  "message": "Answer submitted successfully for review"
}
```

### Handler Implementation

**File**: `/services/tenant-service/internal/handler/client_audit_view.go`

**Key Features:**
- Extracts `client_id` and `user_id` from JWT context
- Determines if user is POC or stakeholder based on `designation` field
- Filters questions based on role (POC sees all, stakeholders see assigned)
- Handles submission creation and updates
- Manages submission status transitions

**Helper Functions:**
- `getClientIDFromUser()` - Extract client ID from JWT
- `getUserIDFromUser()` - Extract user ID from JWT
- `isUserPOCRole()` - Check if user is POC (can see all questions)

## Frontend Implementation

### Routes

**File**: `/apps/frontend/app/routes.ts`

```typescript
route("audit", "routes/audit.tsx"),              // List page
route("audit/:auditId", "routes/audit.$auditId.tsx"),  // Detail page
```

### Pages

#### 1. Audit List Page (`/audit`)

**File**: `/apps/frontend/app/routes/audit.tsx`

**Features:**
- Displays all assigned frameworks as cards
- Shows progress bars for each audit
- Status badges (Not Started, In Progress, Under Review, Completed, Overdue)
- Due date indicators
- Question counts (total vs answered)
- Click to navigate to detail page

**UI Components:**
- Grid layout for audit cards
- Progress bars with color coding
- Status icons and badges
- Empty state for no audits

#### 2. Audit Detail Page (`/audit/:auditId`)

**File**: `/apps/frontend/app/routes/audit.$auditId.tsx`

**Features:**
- Questions grouped by section
- Role-based question visibility (automatic)
- Question type support:
  - **Yes/No**: Radio buttons (Yes, No, N/A)
  - **Text**: Textarea for free-form answers
  - **Multiple Choice**: (Ready for implementation)
- Mandatory field indicators
- Assignment badges ("Assigned to you")
- Submission status badges
- Draft saving functionality
- Submit for review button
- Read-only mode for submitted/approved answers

**Form Handling:**
- Local state management for form data
- Auto-populate existing answers
- Validation (explanation required)
- Save as draft (in_progress status)
- Submit for review (submitted status)

**UI Components:**
- Section-based question grouping
- Question cards with status indicators
- Radio groups for Yes/No questions
- Textareas for text answers and explanations
- Action buttons (Save Draft, Submit for Review)
- Toast notifications for success/error

### API Client

**File**: `/apps/frontend/app/api/client-audit.ts`

```typescript
export const clientAuditApi = {
  listAudits: () => Promise<ClientAudit[]>,
  getAuditDetail: (auditId: string) => Promise<ClientAuditDetail>,
  saveSubmission: (payload: ClientSubmissionPayload) => Promise<any>,
  submitAnswer: (submissionId: string) => Promise<any>,
};
```

### TypeScript Types

**File**: `/apps/frontend/app/types/index.ts`

```typescript
interface ClientAudit {
  id: string;
  framework_id: string;
  framework_name: string;
  due_date: string;
  status: 'not_started' | 'in_progress' | 'under_review' | 'completed' | 'overdue';
  total_questions: number;
  answered_count: number;
  progress_percent: number;
  created_at: string;
}

interface ClientAuditQuestion {
  id: string;
  section: string;
  question_number: string;
  question_text: string;
  question_type: 'yes_no' | 'text' | 'multiple_choice';
  help_text?: string;
  is_mandatory: boolean;
  display_order: number;
  submission_id?: string;
  answer_value?: 'yes' | 'no' | 'na';
  answer_text?: string;
  explanation?: string;
  submission_status?: 'not_started' | 'in_progress' | 'submitted' | 'approved' | 'rejected' | 'referred';
  submitted_at?: string;
  submitted_by?: string;
  is_assigned_to_me: boolean;
}

interface ClientSubmissionPayload {
  question_id: string;
  answer_value?: 'yes' | 'no' | 'na';
  answer_text?: string;
  explanation: string;
}
```

## Role-Based Access Control

### User Roles

1. **poc_client** - Client Point of Contact
   - Can see ALL questions in assigned audits
   - Can answer any question
   - Can assign questions to stakeholders

2. **poc_internal** - Internal Point of Contact
   - Can see ALL questions in assigned audits
   - Can answer any question
   - Can assign questions to stakeholders

3. **stakeholder** - Client Stakeholder
   - Can ONLY see questions assigned to them
   - Can answer assigned questions
   - Cannot see unassigned questions

### Permission Requirements

All client audit endpoints require:
- Valid authentication (JWT cookie)
- `client_id` in user claims
- Appropriate permissions:
  - `audit:list` - View audit list
  - `audit:read` - View audit details
  - `audit:submit` - Submit answers

### Sidebar Configuration

**File**: `/apps/frontend/app/components/layout/Sidebar.tsx`

Added "Audit" menu item:
```typescript
{
  title: 'Audit',
  href: '/audit',
  icon: CircleCheckBig,
  permission: 'audit:list',
}
```

Visible to users with designation: `poc_client`, `poc_internal`, `stakeholder`

## Data Flow

### Viewing Audits

```
1. User navigates to /audit
   ↓
2. Frontend calls GET /api/client-audit
   ↓
3. Backend extracts client_id from JWT
   ↓
4. Backend queries client database for audits
   ↓
5. Backend calculates progress for each audit
   ↓
6. Frontend displays audit cards with progress
```

### Viewing Questions

```
1. User clicks on audit card
   ↓
2. Frontend calls GET /api/client-audit/:auditId
   ↓
3. Backend extracts client_id and user_id from JWT
   ↓
4. Backend checks if user is POC or stakeholder
   ↓
5. Backend queries questions with role-based filtering:
   - POC: All questions
   - Stakeholder: Only assigned questions
   ↓
6. Backend joins with submissions to get existing answers
   ↓
7. Frontend groups questions by section
   ↓
8. Frontend displays questions with form inputs
```

### Submitting Answers

```
1. User fills out question form
   ↓
2. User clicks "Save Draft"
   ↓
3. Frontend calls POST /api/client-audit/submissions
   ↓
4. Backend creates/updates submission with status "in_progress"
   ↓
5. User clicks "Submit for Review"
   ↓
6. Frontend calls POST /api/client-audit/submissions/:id/submit
   ↓
7. Backend updates submission status to "submitted"
   ↓
8. Auditor can now review the submission
```

## Question Types

### 1. Yes/No Questions
- Radio buttons with options: Yes, No, N/A
- Stored in `answer_value` field
- Explanation required

### 2. Text Questions
- Textarea for free-form text
- Stored in `answer_text` field
- Explanation required

### 3. Multiple Choice (Ready for implementation)
- Can be added with radio buttons or checkboxes
- Options stored in question metadata
- Selected option stored in `answer_value`

## Submission States

1. **not_started** - Question not yet answered
2. **in_progress** - Draft saved, not submitted
3. **submitted** - Submitted for auditor review
4. **approved** - Auditor approved the answer
5. **rejected** - Auditor rejected, needs revision
6. **referred** - Referred to another stakeholder

## Security Features

### Authentication
- JWT-based authentication via cookies
- User context includes `client_id`, `user_id`, `designation`
- All endpoints require valid authentication

### Authorization
- Permission-based access control
- Client isolation (users can only access their client's data)
- Role-based question filtering (POC vs stakeholder)

### Data Validation
- Question ID validation
- Explanation required for all submissions
- Status transition validation
- User ownership validation

## Future Enhancements

### 1. Question Assignment UI
- POC users can assign questions to stakeholders
- Bulk assignment functionality
- Assignment notifications

### 2. Evidence Upload
- File upload for supporting documents
- Evidence linked to submissions
- File preview and download

### 3. Comments and Discussion
- Comment threads on questions
- Internal vs external comments
- @mentions for stakeholders

### 4. Submission History
- Version tracking for resubmissions
- Audit trail of changes
- Comparison view for versions

### 5. Notifications
- Email notifications for assignments
- Reminders for pending questions
- Status change notifications

### 6. Bulk Operations
- Save all drafts at once
- Submit multiple answers together
- Export/import answers

### 7. Progress Tracking
- Section-level progress
- Time tracking per question
- Completion estimates

## Testing

### Manual Testing Steps

1. **Login as POC User**:
   - Navigate to `/audit`
   - Verify all assigned frameworks are visible
   - Click on a framework
   - Verify ALL questions are visible
   - Fill out a question and save as draft
   - Submit the answer for review

2. **Login as Stakeholder**:
   - Navigate to `/audit`
   - Verify assigned frameworks are visible
   - Click on a framework
   - Verify ONLY assigned questions are visible
   - Fill out an assigned question
   - Save and submit

3. **Test Question Types**:
   - Yes/No: Select radio button, add explanation
   - Text: Enter text answer, add explanation
   - Verify validation (explanation required)

4. **Test Submission States**:
   - Save as draft (in_progress)
   - Submit for review (submitted)
   - Verify read-only mode for submitted answers

### API Testing

```bash
# List audits
curl -H "Cookie: auth_token=YOUR_TOKEN" \
  http://localhost:8080/api/client-audit

# Get audit detail
curl -H "Cookie: auth_token=YOUR_TOKEN" \
  http://localhost:8080/api/client-audit/{audit_id}

# Save submission
curl -X POST -H "Cookie: auth_token=YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"question_id":"uuid","answer_value":"yes","explanation":"..."}' \
  http://localhost:8080/api/client-audit/submissions

# Submit answer
curl -X POST -H "Cookie: auth_token=YOUR_TOKEN" \
  http://localhost:8080/api/client-audit/submissions/{submission_id}/submit
```

## Summary

✅ **Backend**:
- SQL queries with role-based filtering
- API handlers for list, detail, save, submit
- JWT-based authentication and authorization
- Client database integration
- Submission workflow management

✅ **Frontend**:
- Audit list page with progress tracking
- Audit detail page with question forms
- Role-based question visibility (automatic)
- Draft saving and submission
- TypeScript types and API client
- Responsive UI with status indicators

✅ **Security**:
- Permission-based access control
- Client data isolation
- Role-based question filtering
- Validation and error handling

The client audit module provides a complete solution for clients to view assigned compliance frameworks and submit responses to questions, with proper role-based access control ensuring POC users see all questions while stakeholders only see their assigned questions.
