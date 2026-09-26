# Shared Task Prompt Draft: Small 3D Solar System

Status: Draft for review. Do not submit this document as a benchmark prompt.

The task text below contains the requirements accepted in ADRs 0003 through 0005. Before freezing an executable prompt, agree the exact Three.js release and delivery, artifact submission format, and concrete execution conditions. The final prompt will include those agreed constraints and exclude these review notes.

## Proposed Task Text

Create an HTML-based interactive 3D solar system using Three.js.

Requirements:

1. Display one sun and exactly three planets in a 3D scene.
2. Start the planets orbiting the sun automatically when the scene opens.
3. Give the three planets different sizes, colors, and orbital speeds.
4. Provide a pause/resume button that stops and resumes the planets' orbital motion.
5. Provide a global speed control that changes all orbital speeds while preserving their relative differences.

The intended experience is a simple scene whose automatic motion and two controls make its behavior easy to observe.
