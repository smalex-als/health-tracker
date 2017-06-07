package com.tracker.shared;

import java.util.List;
import java.util.Map;

import com.googlecode.jatl.client.HtmlWriter;

public class TableTemplate extends BaseTemplates {

  public HtmlWriter renderPagination(final Map<String, Object> map) {
    return new HtmlWriter() {
      @Override
      protected void build() {
        indent(indentOff);
        renderPagination(this, map);
      }};
  }

  private void renderPagination(HtmlWriter html, final Map<String, Object> map) {
    html.div().classAttr("container");
    html.div().classAttr("row justify-content-center");

    html.start("nav");
    html.ul().classAttr("pagination pagination-lg");
    html.li().classAttr("page-item page-prev").a().classAttr("page-link").href("#").text("<< 2017-03-12").end().end();
    html.li().classAttr("page-item page-current disabled").a().classAttr("page-link").href("#").text("2017-03-17").end().end();
    html.li().classAttr("page-item page-next").a().classAttr("page-link").href("#").text("2017-03-24 >>").end().end();
    html.end();
    html.end();
    html.end();

    html.div().classAttr("row justify-content-center");
    html.div().classAttr("meals-stats").text("").end();
    html.end();

    html.end();
  }

  public HtmlWriter renderListBody(final Map<String, Object> map) {
    return new HtmlWriter() {
      @Override
      protected void build() {
        indent(indentOff);
        renderListBody(this, map);
      }};
  }

  private void renderListBody(HtmlWriter html, final Map<String, Object> map) {
    html.div().classAttr("container");

    html.div().classAttr("search-component");
    html.end();

    html.div().classAttr("row");
    html.div().classAttr("btn-toolbar mb-3");
      html.div().classAttr("btn-group mr-2").attr("role", "group").attr("aria-label", "");
        html.a().classAttr("btn btn-secondary list-btn-add").attr("role", "button").text("Add").end();
        html.a().classAttr("btn btn-secondary list-btn-rm").attr("role", "button").text("Delete").end();
      html.end();

      // html.div().classAttr("input-group");
      //   html.input().type("text").classAttr("form-control search-input").attr("placeholder", "Search for...").end();
      //   html.span().classAttr("input-group-btn");
      //     html.button().classAttr("btn btn-secondary search-btn").type("button").text("Go!").end();
      //   html.end();
      // html.end();

    html.end();
    html.end();


    html.end();
  }

  public HtmlWriter renderTable(final Map<String, Object> map) {
    return new HtmlWriter() {
      @Override
      protected void build() {
        indent(indentOff);
        printRows(this, map);
      }};
  }

  private void printRows(HtmlWriter html, final Map<String, Object> map) {
    List<Map<String, Object>> rows = (List<Map<String, Object>>) map.get("rows");
    if (rows == null || rows.size() == 0) {
      html.div().classAttr("row no-records-found");
      html.h4().text("No records found!").end();
      html.end();
      return;
    }

    List<Map<String, Object>> columns = getList(map, "columns");
    html.div().classAttr("row");
    html.div().classAttr("table-responsive");
    html.table().classAttr("table table-striped");

    // head start
    html.thead();
    html.tr();
    html.th().text("*").end();
    for (Map<String, Object> column : columns) {
      html.th();
      String align = getString(column, "align");
      if (align != null && align.length() > 0) {
        // we need to override th style
        html.style("text-align:" + align + ";");
      }
      String width = getString(column, "width");
      if (width != null && width.length() > 0) {
        html.width(width);
      }
      html.text(getString(column, "id")).end();
    }
    html.end();
    html.end();
    // head end

    html.tbody();
    if (rows != null) {
      for (Map<String, Object> row : rows) {
        List<String> values = (List<String>) row.get("values");
        html.tr();
        html.attr("data-id", getString(row, "id"));
        html.td().input().type("checkbox").end().end();
        int col = 0;
        for (String value : values) {
          html.td();
          String align = getString(columns.get(col), "align");
          if (align != null && align.length() > 0) {
            html.align(align);
          }
          html.text(value).end();
          col++;
        }
        html.end();
      }
    }
    html.end(); // tbody

    html.end(); // table
    html.end(); // table-responsive
    html.end(); // row
  }

  private void renderSearch(HtmlWriter html, final Map<String, Object> map) {
    html.div();

    html.form().classAttr("search-form");
    if (map.containsKey("dateSearch")) {
      html.div().classAttr("form-group row");
      html.label().classAttr("col-form-label").text("Date within").end();
      html.raw("&nbsp;");
      html.div().classAttr("");
      html.select().classAttr("custom-select search-with-days");
      html.option().value("1").text("1 day").end();
      html.option().value("3").text("3 days").end();
      html.option().value("7").text("1 week").end();
      html.option().value("14").text("2 weeks").end();
      html.option().value("30").text("1 month").end();
      html.option().value("60").text("2 months").end();
      html.end();
      html.end(); 

      html.raw("&nbsp;");
      html.label().classAttr("col-form-label").text("of").end();
      html.raw("&nbsp;");
      html.div().classAttr("");
      html.input().classAttr("form-control search-bydate").type("date").end();
      html.end(); 
      html.end(); // row
    }
    html.div().classAttr("form-group row");
    html.label().classAttr("col-form-label").text("Contains words").end();
    html.raw("&nbsp;");
    html.div().classAttr("");
    html.input().type("text").classAttr("search-input form-control").maxlength("64").attr("placeholder", "").end();
    html.end();
    html.end();

    if (map.containsKey("dateSearch")) {
      html.div().classAttr("form-group row");

      html.label().classAttr("col-form-label").text("Time from").end();
      html.raw("&nbsp;");
      html.div().classAttr("");
      html.input().classAttr("form-control search-fromtime").type("time").end();
      html.end();

      html.raw("&nbsp;");
      html.label().classAttr("col-form-label").text("to").end();
      html.raw("&nbsp;");
      html.div().classAttr("");
      html.input().classAttr("form-control search-totime").type("time").end();
      html.end();

      html.end(); // end
    }
    html.button().classAttr("btn btn-primary").text("Search").end();

    html.end(); // form


    html.end(); // row
  }

  public HtmlWriter renderSearch(Map<String, Object> map) {
    return new HtmlWriter() {
      @Override
      protected void build() {
        indent(indentOff);
        renderSearch(this, map);
      }};
  }
}
