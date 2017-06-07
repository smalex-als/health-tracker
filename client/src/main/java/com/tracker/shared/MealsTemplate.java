package com.tracker.shared;

import java.util.List;
import java.util.Map;

import com.googlecode.jatl.client.HtmlWriter;

public class MealsTemplate extends BaseTemplates {
  public HtmlWriter renderStats(final Map<String, Object> map) {
    return new HtmlWriter() {
      @Override
      protected void build() {
        indent(indentOff);
        renderStats(this, map);
      }};
  }

  private void renderStats(HtmlWriter html, final Map<String, Object> map) {
    List<Map<String, Object>> rows = getList(map, "items");
    if (rows == null || rows.size() == 0) {
      html.div().classAttr("row no-records-found");
      html.h4().text("No records found!").end();
      html.end();
      return;
    }

    html.div().classAttr("row");
    html.div().classAttr("table-responsive");
    html.table().classAttr("table table-striped");

    // head start
    // html.thead();
    // html.tr();
    //   html.th().text("date").end();
    //   html.th().text("total").end();
    //   html.th().text("success").end();
    // html.end();
    // html.end();
    // head end

    html.tbody();
    html.tr();
    for (Map<String, Object> row : rows) {
      html.td().text(getString(row, "weekday")).end();
    }
    html.end();

    html.tr();
    for (Map<String, Object> row : rows) {
      html.td().align("right");
      if (getBoolean(row, "success")) {
        html.style("background-color: #8FBC8F;");
      } else {
        html.style("background-color: #FFB6C1;");
      }
      int val = getInteger(row, "total");
      if (val != 0) {
        html.a().href("#l:meals/date=" + getString(row, "date") + "&days=1");
        html.text(String.valueOf(val));
        html.end();
      } else {
        html.raw("&nbsp;");
      }
      html.end();
    }
    html.end(); // tr

    html.end(); // tbody

    html.end(); // table
    html.end(); // table-responsive
    html.end(); // row
  }}
