import json
import sys
import os

def create_svg(filename, title, y_label, points_dict, color):
    width, height, padding = 800, 400, 60
    
    # Находим максимальное значение для корректного масштаба оси Y
    values = [v for v in points_dict.values() if isinstance(v, (int, float))]
    max_val = max(values) * 1.2 if values and max(values) > 0 else 100.0
    min_val = 0
    
    svg = f'''<?xml version="1.0" encoding="UTF-8"?>
<svg width="{width}" height="{height}" xmlns="http://www.w3.org/2000/svg">
  <rect width="100%" height="100%" fill="#f8f9fa" rx="8"/>
  <text x="{width/2}" y="30" text-anchor="middle" font-family="Arial, sans-serif" font-size="18" font-weight="bold" fill="#2c3e50">{title}</text>
  <line x1="{padding}" y1="{height-padding}" x2="{width-padding}" y2="{height-padding}" stroke="#34495e" stroke-width="2"/>
  <line x1="{padding}" y1="{padding}" x2="{padding}" y2="{height-padding}" stroke="#34495e" stroke-width="2"/>
  <text x="20" y="{height/2}" text-anchor="middle" font-family="Arial, sans-serif" font-size="14" fill="#34495e" transform="rotate(-90, 20, {height/2})">{y_label}</text>'''
    
    # Рисуем метки оси Y
    for i in range(5):
        y_val = min_val + (max_val - min_val) * i / 4
        y_pos = height - padding - (height - 2*padding) * i / 4
        svg += f'<text x="{padding-10}" y="{y_pos+5}" text-anchor="end" font-family="Arial, sans-serif" font-size="12" fill="#7f8c8d">{y_val:.0f}</text>'
        svg += f'<line x1="{padding-5}" y1="{y_pos}" x2="{padding}" y2="{y_pos}" stroke="#95a5a6" stroke-width="1"/>'
        
    # Рисуем точки и линии
    keys = list(points_dict.keys())
    x_step = (width - 2*padding) / (len(keys) - 1) if len(keys) > 1 else 0
    y_scale = (height - 2*padding) / (max_val - min_val) if (max_val - min_val) > 0 else 1
    
    path_d = ""
    for i, key in enumerate(keys):
        val = float(points_dict[key])
        x = padding + i * x_step
        y = height - padding - (val - min_val) * y_scale
        
        if i == 0: 
            path_d += f"M {x},{y}"
        else: 
            path_d += f" L {x},{y}"
        
        svg += f'''
  <circle cx="{x}" cy="{y}" r="5" fill="{color}" stroke="#fff" stroke-width="2"/>
  <text x="{x}" y="{y-15}" text-anchor="middle" font-family="Arial, sans-serif" font-size="12" font-weight="bold" fill="{color}">{val:.1f}</text>
  <text x="{x}" y="{height-padding+20}" text-anchor="middle" font-family="Arial, sans-serif" font-size="11" fill="#7f8c8d">{key}</text>'''
        
    if len(keys) > 1:
        svg += f'\n  <path d="{path_d}" stroke="{color}" stroke-width="3" fill="none"/>'
    svg += "\n</svg>"
    
    with open(filename, 'w', encoding='utf-8') as f:
        f.write(svg)
    print(f"✅ Создан файл: {filename}")

def main():
    if len(sys.argv) != 2:
        print("Использование: python3 generate_charts.py <summary.json>")
        return
        
    json_file = sys.argv[1]
    base_name = os.path.splitext(json_file)[0]
    
    with open(json_file, 'r', encoding='utf-8') as f:
        data = json.load(f)
        
    metrics = data.get('metrics', {})
    
    # 1. График Latency (теперь читаем напрямую, без .get('values', {}))
    req_dur = metrics.get('http_req_duration', {})
    latency_data = {
        "avg": req_dur.get('avg', 0),
        "med": req_dur.get('med', 0),
        "p(90)": req_dur.get('p(90)', 0),
        "p(95)": req_dur.get('p(95)', 0)
    }
    print(f"📊 Читаем Latency: {latency_data}")
    create_svg(f"{base_name}_latency.svg", f"Latency (время отклика) - {os.path.basename(json_file)}", "Время (мс)", latency_data, "#e74c3c")
    
    # 2. График Throughput (читаем rate напрямую)
    reqs = metrics.get('http_reqs', {})
    rps = reqs.get('rate', 0)
    throughput_data = {"RPS": rps}
    print(f"📊 Читаем Throughput: {throughput_data}")
    create_svg(f"{base_name}_throughput.svg", f"Throughput (RPS) - {os.path.basename(json_file)}", "Запросов/сек", throughput_data, "#27ae60")

if __name__ == "__main__":
    main()