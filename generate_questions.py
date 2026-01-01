#!/usr/bin/env python3
"""
Script to generate questions.json from questions_text.txt and answer key.
Run this script to update the questions.json file with all questions.
"""

import json
import re

# Answer key mapping (a=1, b=2, c=3, d=4)
answers = {
    1: 1, 2: 1, 3: 2, 4: 1, 5: 1, 6: 1, 7: 3, 8: 3, 9: 3, 10: 2,
    11: 2, 12: 1, 13: 3, 14: 2, 15: 3, 16: 2, 17: 1, 18: 1, 19: 2, 20: 2,
    21: 1, 22: 4, 23: 1, 24: 3, 25: 3, 26: 3, 27: 2, 28: 1, 29: 2, 30: 1,
    31: 3, 32: 4, 33: 1, 34: 4, 35: 2, 36: 3, 37: 3, 38: 1, 39: 1, 40: 2,
    41: 4, 42: 1, 43: 3, 44: 2, 45: 1, 46: 1, 47: 4, 48: 2, 49: 4, 50: 2,
    51: 4, 52: 4, 53: 1, 54: 4, 55: 3, 56: 1, 57: 1, 58: 3, 59: 2, 60: 1,
    61: 2, 62: 2, 63: 3, 64: 1, 65: 4, 66: 3, 67: 3, 68: 4, 69: 2, 70: 2,
    71: 3, 72: 2, 73: 1, 74: 2, 75: 1, 76: 1, 77: 1, 78: 3, 79: 2, 80: 1,
    81: 3, 82: 4, 83: 4, 84: 4, 85: 1, 86: 3, 87: 3, 88: 2, 89: 1, 90: 3,
    91: 3, 92: 1, 93: 2, 94: 1, 95: 2, 96: 4, 97: 1, 98: 3, 99: 4, 100: 3,
    101: 2, 102: 3, 103: 3, 104: 1, 105: 4, 106: 3, 107: 4, 108: 3, 109: 2, 110: 1,
    111: 2, 112: 4, 113: 3, 114: 3, 115: 2, 116: 4,
}

def parse_questions():
    """Parse questions from questions_text.txt file."""
    with open('questions_text.txt', 'r', encoding='utf-8') as f:
        content = f.read()
    
    json_questions = []
    
    # Split by question numbers
    parts = re.split(r'^(\d+)\s+', content, flags=re.MULTILINE)
    
    for i in range(1, len(parts), 2):
        if i + 1 >= len(parts):
            break
        q_num = int(parts[i])
        q_content = parts[i + 1]
        
        # Split question text and answers
        lines = q_content.split('\n')
        q_text_lines = []
        answer_lines = []
        
        for line in lines:
            line = line.strip()
            if not line or '©' in line or 'face2face' in line.lower() or 'Photocopiable' in line:
                continue
            if re.match(r'^[a-d]\)', line):
                answer_lines.append(line)
            else:
                q_text_lines.append(line)
        
        # Parse answers
        answer_dict = {}
        for ans_line in answer_lines:
            match = re.match(r'^([a-d])\)\s*(.+)$', ans_line)
            if match:
                letter = match.group(1)
                text = match.group(2).strip()
                num = ord(letter) - ord('a') + 1
                answer_dict[num] = text
        
        if len(answer_dict) >= 3 and q_text_lines:
            # Clean question text
            q_text = ' '.join(q_text_lines).strip()
            # Handle dialogue format (names on separate lines) - only for first few questions
            if q_num <= 2:
                # Preserve dialogue format
                q_text = '\n'.join(q_text_lines).strip()
            else:
                # For other questions, clean up but preserve intentional line breaks
                # Handle question marks and periods that should be on new lines
                q_text = re.sub(r'([?!])\s+([A-Z])', r'\1\n\2', q_text)
                # Clean up multiple spaces
                q_text = re.sub(r' +', ' ', q_text)
                # Clean up multiple newlines
                q_text = re.sub(r'\n+', '\n', q_text).strip()
            
            json_questions.append({
                "text": q_text,
                "text_html": f"<b>Question {q_num}</b>\n\n{q_text}",
                "answer_1": answer_dict.get(1, ''),
                "answer_1_html": f"1. {answer_dict.get(1, '')}",
                "answer_2": answer_dict.get(2, ''),
                "answer_2_html": f"2. {answer_dict.get(2, '')}",
                "answer_3": answer_dict.get(3, ''),
                "answer_3_html": f"3. {answer_dict.get(3, '')}",
                "answer_4": answer_dict.get(4, '') if 4 in answer_dict else "",
                "answer_4_html": f"4. {answer_dict.get(4, '')}" if 4 in answer_dict else "",
                "correct_answer_id": answers.get(q_num, 1),
                "score": 1
            })
    
    return json_questions

if __name__ == '__main__':
    json_questions = parse_questions()
    
    # Write to JSON
    questions_data = {"questions": json_questions}
    with open('questions.json', 'w', encoding='utf-8') as f:
        json.dump(questions_data, f, indent=2, ensure_ascii=False)
    
    print(f"Generated questions.json with {len(json_questions)} questions")

