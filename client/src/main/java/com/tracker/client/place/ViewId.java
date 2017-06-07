package com.tracker.client.place;

public class ViewId {
  private String activityId = "";
  private String fieldId = "";

  public ViewId(String activityId) {
    this(activityId, null);
  }

  public ViewId(String activityId, String fieldId) {
    this.activityId = activityId;
    this.fieldId = fieldId;
  }

  public String getActivityId() {
    return activityId;
  }

  public String getFieldId() {
    return fieldId;
  }

  @Override
  public boolean equals(Object o) {
    if (this == o) return true;
    if (o == null || getClass() != o.getClass()) return false;

    ViewId viewId = (ViewId) o;

    if (activityId != null ? !activityId.equals(viewId.activityId) : viewId.activityId != null) return false;
    if (fieldId != null ? !fieldId.equals(viewId.fieldId) : viewId.fieldId != null) return false;

    return true;
  }

  @Override
  public int hashCode() {
    int result = (activityId != null ? activityId.hashCode() : 0);
    result = 31 * result + (fieldId != null ? fieldId.hashCode() : 0);
    return result;
  }

  @Override
  public String toString() {
    return "ViewId{" +
        "activityId='" + activityId + '\'' +
        ", fieldId='" + fieldId + '\'' +
        '}';
  }
}
