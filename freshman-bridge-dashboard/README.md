# Freshman Analytics Dashboard

Welcome to the Freshman Analytics Dashboard! This application is designed as an EdTech platform tailored for first-year university students. Its goal is to reduce academic stress and facilitate a smoother transition into university life by providing personalized progress stats, actionable micro-tasks, targeted gap analysis, and an integrated AI tutor.

## Getting Started

To view and interact with the dashboard:

1. Open `index.html` in your web browser. No complex server setup is required as it uses standard HTML, CSS (Tailwind via CDN), and Vanilla JavaScript.
2. If you are using an IDE like VS Code, you can use the "Live Server" extension, or simply double-click the `index.html` file in your file explorer.

## Outline of Dashboard Components

The dashboard is divided into several intuitive sections to help you manage your learning effectively:

### 1. Left Sidebar Navigation
The sidebar provides quick access to different areas of the platform:
- **Dashboard**: Your main overview (current view).
- **Courses**: Access your enrolled subjects and course materials.
- **Schedule**: View your upcoming classes and deadlines.
- **Analytics**: Deep dive into your learning statistics.
- **Settings**: Configure your profile and preferences.

### 2. Top Header
The header gives you a personalized greeting and quick actions:
- **Greeting**: Personalized welcome message based on the time of day.
- **Notifications**: Alerts for upcoming deadlines or new materials.
- **Add Material Button**: Quick action to upload new study resources or assignments.
- **Profile Avatar**: Access to your account settings.

### 3. Quick Stats (Row 1)
At-a-glance metrics to track your overall performance:
- **Completion Rate**: Shows the percentage of tasks and courses you've finished.
- **On-Time Rate**: Tracks how often you submit assignments before the deadline.
- **Avg. Completion Time**: Displays the average time spent on tasks or modules.

### 4. Action & Diagnosis (Row 2)
Actionable insights to focus your study sessions:
- **Today's Micro-Tasks**: A checklist of small, manageable tasks (e.g., "Memorize 5 key definitions"). You can check these off as you complete them to maintain momentum.
- **Targeted Gap Analysis**: A radar chart visualizing your proficiency in specific topics (e.g., Demand, Supply, Elasticity). It highlights areas where you need improvement and provides a "Start Focused Review" button.

### 5. Path Optimization (Row 3)
A visualization of your learning journey:
- **Optimal Learning Path Chart**: Maps task difficulty against your skill level to keep you in the "Optimal Zone" (Flow state), avoiding both "Boredom" (too easy) and "Stress" (too hard).

### 6. Intelligent AI Partner (Chat Widget)
A floating chat widget in the bottom right corner:
- Acts as a 24/7 tutor.
- Context-aware: It analyzes your current tasks and materials (e.g., offering a quick quiz on "Elasticity of Demand" if that's your identified gap).
- Interactive: You can type messages and receive automated, helpful responses to guide your studying.

## Technologies Used
- **HTML5**: For semantic structure.
- **Tailwind CSS**: For rapid, modern, and responsive styling (loaded via CDN).
- **Vanilla JavaScript**: For interactive elements like task checking, navigation highlighting, toast notifications, and the chat widget logic.
- **Font Awesome**: For scalable vector icons.
- **Google Fonts (Inter)**: For clean, readable typography.
